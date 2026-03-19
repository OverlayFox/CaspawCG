package ui

import (
	"caspaw-cg/src/data"
	"caspaw-cg/src/types"
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
