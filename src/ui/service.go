package ui

import (
	"github.com/overlayfox/caspaw-cg/src/data"
	"github.com/overlayfox/caspaw-cg/src/types"
)

// UIService bridges the UI with the GoLang system
type UIService struct {
	app               *App
	datasourceManager data.DatasourceManager
	casparCGClients   []types.CasparCGClient
}

func NewUIService(app *App, datasourceManager data.DatasourceManager, casparCGClients []types.CasparCGClient) *UIService {
	return &UIService{
		app:               app,
		datasourceManager: datasourceManager,
		casparCGClients:   casparCGClients,
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

func (u *UIService) PushCasparCGData(template string, layer, channel int, data map[string]any, sizing types.Sizing) {
	for _, client := range u.casparCGClients {
		err := client.PushCGData(template, layer, channel, data, sizing)
		if err != nil {
			u.app.logger.Error().Err(err).Msgf("Failed to push CG data to template '%s' on layer %d, channel %d", template, layer, channel)
		}
	}
}

func (u *UIService) StopCasparCGData(template string, layer, channel int) {
	for _, client := range u.casparCGClients {
		err := client.StopCGData(template, layer, channel)
		if err != nil {
			u.app.logger.Error().Err(err).Msgf("Failed to stop CG data for template '%s' on layer %d, channel %d", template, layer, channel)
		}
	}
}

func (u *UIService) PrimeDataSource(name string, locations []data.Location) error {
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

func (u *UIService) GetDataSourceValue(name string, location data.Location) (data.Data, error) {
	ds, err := u.datasourceManager.GetDataSource(name)
	if err != nil {
		u.app.logger.Error().Err(err).Msgf("Failed to get datasource '%s'", name)
		return data.Data{}, err
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
	Channel  int
	Data     map[string]any
	Sizing   types.Sizing
}

func (u *UIService) PushCasparCGDataGroup(dataGroups []CGDataGroup) {
	for _, data := range dataGroups {
		u.PushCasparCGData(data.Template, data.Layer, data.Channel, data.Data, data.Sizing)
	}
}

func (u *UIService) StopCasparCGDataGroup(dataGroups []CGDataGroup) {
	for _, data := range dataGroups {
		u.StopCasparCGData(data.Template, data.Layer, data.Channel)
	}
}
