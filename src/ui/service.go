package ui

import (
	"context"
	"sync"
	"time"

	"github.com/overlayfox/casparcg-amcp-go/types/responses"

	"github.com/overlayfox/caspaw-cg/src/types"
)

// UIService bridges the UI with the GoLang system
type UIService struct {
	app               *App
	datasourceManager types.DatasourceManager
	casparCGClients   []types.CasparCGClient
	updateHandler     *UpdateHandler

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewUIService(upstreamCtx context.Context, app *App, datasourceManager types.DatasourceManager, casparCGClients []types.CasparCGClient) *UIService {
	ctx, cancel := context.WithCancel(upstreamCtx)
	return &UIService{
		app:               app,
		datasourceManager: datasourceManager,
		casparCGClients:   casparCGClients,
		updateHandler:     NewUpdateHandler(ctx, app.logger, datasourceManager, casparCGClients),
		ctx:               ctx,
		cancel:            cancel,
	}
}

func (u *UIService) SaveLayout(config LayoutConfig) error {
	u.app.logger.Info().Msg("Saving layout configuration")
	return SaveLayout(config)
}

func (u *UIService) LoadLayout() (LayoutConfig, error) {
	u.app.logger.Info().Msg("Loading layout configuration")
	return LoadLayout()
}

func (u *UIService) GetDataSources() []string {
	names := u.datasourceManager.GetDataSourceNames()
	if len(names) == 0 {
		return []string{
			"No datasources available",
		}
	}
	return names
}

func (u *UIService) GetCasparCGTemplates() []string {
	result := make([]string, 0)
	for _, client := range u.casparCGClients {
		templates, err := client.GetTemplates()
		if err != nil {
			u.app.logger.Error().Err(err).Msg("Failed to get templates from CasparCG client")
			continue
		}
		result = append(result, templates...)
	}
	return result
}

func (u *UIService) GetCasparCGMedia() []string {
	result := make([]string, 0)
	for _, client := range u.casparCGClients {
		media, err := client.GetMedia()
		if err != nil {
			u.app.logger.Error().Err(err).Msg("Failed to get media from CasparCG client")
			continue
		}
		result = append(result, media...)
	}
	return result
}

func (u *UIService) GetCasparCGMediaInfo(filename string) (responses.CINF, error) {
	for _, client := range u.casparCGClients {
		info, err := client.GetMediaInfo(filename)
		if err != nil {
			u.app.logger.Error().Err(err).Msgf("Failed to get media info for '%s' from CasparCG client", filename)
			continue
		}
		return info, nil
	}
	return responses.CINF{}, nil
}

func (u *UIService) PushCasparCGData(template string, layer int, channels []int, data map[string]any, sizing types.Sizing, delay time.Duration) {
	for _, client := range u.casparCGClients {
		u.wg.Add(1)
		go func(client types.CasparCGClient) {
			defer u.wg.Done()
			err := client.AddCGData(template, layer, channels, data, sizing, delay)
			if err != nil {
				u.app.logger.Error().Err(err).Msgf("Failed to push CG data to template '%s' on layer %d, channels %v", template, layer, channels)
			}
		}(client)
	}
}

func (u *UIService) StopCasparCGData(template string, layer int, channels []int, delay time.Duration) {
	for _, client := range u.casparCGClients {
		u.wg.Add(1)
		go func(client types.CasparCGClient) {
			defer u.wg.Done()
			err := client.StopCGData(template, layer, channels, delay)
			if err != nil {
				u.app.logger.Error().Err(err).Msgf("Failed to stop CG data for template '%s' on layer %d, channels %v", template, layer, channels)
			}
		}(client)
	}
}

func (u *UIService) NextCasparCGData(template string, layer int, channels []int, delay time.Duration) {
	for _, client := range u.casparCGClients {
		u.wg.Add(1)
		go func(client types.CasparCGClient) {
			defer u.wg.Done()
			err := client.NextCGData(template, layer, channels, delay)
			if err != nil {
				u.app.logger.Error().Err(err).Msgf("Failed to go to next CG data for template '%s' on layer %d, channels %v", template, layer, channels)
			}
		}(client)
	}
}

type UpdateData struct {
	UID       int
	CasparKey string
	Type      string
	Range     string
	Offset    string
}

func (u *UIService) PrimeDataSource(name string, locations []types.Location) error {
	ds, err := u.datasourceManager.GetDataSource(name)
	if err != nil {
		u.app.logger.Error().Err(err).Msgf("Failed to get datasource '%s'", name)
		return err
	}

	u.app.logger.Info().Msgf("Priming datasource '%s' with locations: %v", name, locations)
	if err := ds.Prime(locations); err != nil { // TODO: add removal of primed data when a new PrimDataSource is called
		u.app.logger.Error().Err(err).Msgf("Failed to prime datasource '%s'", name)
		return err
	}
	return nil
}

func (u *UIService) GetDataSourceValue(name string, location types.Location) (types.Data, error) {
	ds, err := u.datasourceManager.GetDataSource(name)
	if err != nil {
		u.app.logger.Error().Err(err).Msgf("Failed to get datasource '%s'", name)
		return types.Data{}, err
	}

	u.app.logger.Info().Msgf("Getting value from datasource '%s' for location: %v", name, location)
	data, err := ds.Get(location.Key)
	if err != nil {
		u.app.logger.Error().Err(err).Msgf("Failed to get value from datasource '%s' for location: %v", name, location)
		return data, err
	}
	return data, nil
}

type CGDataGroup struct {
	Template string
	Layer    int
	Channels []int
	Data     map[string]any
	Sizing   types.Sizing
	Delay    time.Duration
}

func (u *UIService) PushCasparCGDataGroup(dataGroups []CGDataGroup) {
	for _, data := range dataGroups {
		u.PushCasparCGData(data.Template, data.Layer, data.Channels, data.Data, data.Sizing, data.Delay)
	}
}

func (u *UIService) StopCasparCGDataGroup(dataGroups []CGDataGroup) {
	for _, data := range dataGroups {
		u.StopCasparCGData(data.Template, data.Layer, data.Channels, data.Delay)
	}
}

func (u *UIService) PlayCasparCGMedia(filename string, layer int, channels []int, loop bool, delay time.Duration) {
	for _, client := range u.casparCGClients {
		u.wg.Add(1)
		go func(client types.CasparCGClient) {
			defer u.wg.Done()
			err := client.PlayMedia(filename, layer, channels, loop, delay)
			if err != nil {
				u.app.logger.Error().Err(err).Msgf("Failed to play media '%s' on layer %d, channels %v", filename, layer, channels)
			}
		}(client)
	}
}

func (u *UIService) StopCasparCGMedia(layer int, channels []int, delay time.Duration) {
	for _, client := range u.casparCGClients {
		u.wg.Add(1)
		go func(client types.CasparCGClient) {
			defer u.wg.Done()
			err := client.StopMedia(layer, channels, delay)
			if err != nil {
				u.app.logger.Error().Err(err).Msgf("Failed to stop media on layer %d, channels %v", layer, channels)
			}
		}(client)
	}
}

func (u *UIService) ClearChannels(channels []int) {
	for _, client := range u.casparCGClients {
		u.wg.Add(1)
		go func(client types.CasparCGClient) {
			defer u.wg.Done()
			client.ClearChannels(channels)
		}(client)
	}
}

func (u *UIService) ClearAll() {
	for _, client := range u.casparCGClients {
		u.wg.Add(1)
		go func(client types.CasparCGClient) {
			defer u.wg.Done()
			client.ClearAll()
		}(client)
	}
}

func (u *UIService) Close() {
	u.cancel()
	u.wg.Wait()
}
