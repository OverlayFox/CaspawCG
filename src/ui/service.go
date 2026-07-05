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
	casparCGClient    types.CasparCGClient
	updateHandler     *UpdateHandler

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewUIService(upstreamCtx context.Context, app *App, datasourceManager types.DatasourceManager, casparCGClients types.CasparCGClient) *UIService {
	ctx, cancel := context.WithCancel(upstreamCtx)
	return &UIService{
		app:               app,
		datasourceManager: datasourceManager,
		casparCGClient:    casparCGClients,
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
	templates, err := u.casparCGClient.GetTemplates()
	if err != nil {
		u.app.logger.Error().Err(err).Msg("Failed to get templates from CasparCG client")
		return nil
	}
	return templates
}

func (u *UIService) GetCasparCGMedia() []string {
	media, err := u.casparCGClient.GetMedia()
	if err != nil {
		u.app.logger.Error().Err(err).Msg("Failed to get media from CasparCG client")
		return nil
	}
	return media
}

func (u *UIService) GetCasparCGMediaInfo(filename string) (responses.CINF, error) {
	info, err := u.casparCGClient.GetMediaInfo(filename)
	if err != nil {
		u.app.logger.Error().Err(err).Msgf("Failed to get media info for '%s' from CasparCG client", filename)
		return responses.CINF{}, err
	}
	return info, nil
}

func (u *UIService) PushCasparCGData(template string, layer int, channels []int, data map[string]any, sizing types.Sizing, delay time.Duration) {
	u.wg.Go(func() {
		err := u.casparCGClient.AddCGData(template, layer, channels, data, sizing, delay)
		if err != nil {
			u.app.logger.Error().Err(err).Msgf("Failed to push CG data to template '%s' on layer %d, channels %v", template, layer, channels)
		}
	})
}

func (u *UIService) StopCasparCGData(template string, layer int, channels []int, delay time.Duration) {
	u.wg.Go(func() {
		err := u.casparCGClient.StopCGData(template, layer, channels, delay)
		if err != nil {
			u.app.logger.Error().Err(err).Msgf("Failed to stop CG data for template '%s' on layer %d, channels %v", template, layer, channels)
		}
	})
}

func (u *UIService) NextCasparCGData(template string, layer int, channels []int, delay time.Duration) {
	u.wg.Go(func() {
		err := u.casparCGClient.NextCGData(template, layer, channels, delay)
		if err != nil {
			u.app.logger.Error().Err(err).Msgf("Failed to go to next CG data for template '%s' on layer %d, channels %v", template, layer, channels)
		}
	})
}

func (u *UIService) UpdateCasparCGData(template string, layer int, channels []int, data map[string]any, delay time.Duration) {
	// push templates live
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
	u.wg.Go(func() {
		err := u.casparCGClient.PlayMedia(filename, layer, channels, loop, delay)
		if err != nil {
			u.app.logger.Error().Err(err).Msgf("Failed to play media '%s' on layer %d, channels %v", filename, layer, channels)
		}
	})
}

func (u *UIService) StopCasparCGMedia(layer int, channels []int, delay time.Duration) {
	u.wg.Go(func() {
		err := u.casparCGClient.StopMedia(layer, channels, delay)
		if err != nil {
			u.app.logger.Error().Err(err).Msgf("Failed to stop media on layer %d, channels %v", layer, channels)
		}
	})
}

func (u *UIService) ClearChannels(channels []int) {
	u.wg.Go(func() {
		u.casparCGClient.ClearChannels(channels)
	})
}

func (u *UIService) ClearAll() {
	u.wg.Go(func() {
		u.casparCGClient.ClearAll()
	})
}

func (u *UIService) Close() {
	u.cancel()
	u.wg.Wait()
}
