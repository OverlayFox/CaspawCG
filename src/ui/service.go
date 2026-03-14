package ui

import "caspaw-cg/src/types"

// UIService bridges the UI with the GoLang system
type UIService struct {
	app               *App
	datasourceManager types.DatasourceManager
}

func NewUIService(app *App, datasourceManager types.DatasourceManager) *UIService {
	return &UIService{app: app, datasourceManager: datasourceManager}
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
