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
