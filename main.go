package main

import (
	"embed"

	"caspaw-cg/src/data"
	"caspaw-cg/src/ui"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := ui.NewApp()
	manager := data.NewManager()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "caspaw-cg",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.Startup,
		Bind: []any{
			app,
			ui.NewUIService(app, manager),
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
