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
