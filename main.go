package main

import (
	"embed"
	"os"

	"caspaw-cg/src/config"
	"caspaw-cg/src/data"
	"caspaw-cg/src/ui"

	"github.com/rs/zerolog"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	cfg, err := config.LoadConfig(logger.With().Str("component", "config").Logger())
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load config")
	}

	app := ui.NewApp()
	dsManager := data.NewManager(cfg.DataSourceManager)

	// Create application with options
	err = wails.Run(&options.App{
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
			ui.NewUIService(app, dsManager),
		},
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to start application")
	}
}
