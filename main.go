package main

import (
	"context"
	"embed"
	"os"

	"github.com/rs/zerolog"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/overlayfox/caspaw-cg/src/config"
	"github.com/overlayfox/caspaw-cg/src/ui"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	var err error
	cfg, err := config.LoadConfig(logger.With().Str("component", "config").Logger())
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load config")
	}
	logger.Info().Msg("Config loaded successfully")

	app, err := ui.NewApp(context.Background(), logger, cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create app")
	}

	err = wails.Run(&options.App{
		Title:    "caspaw-cg",
		Width:    1920,
		Height:   1080,
		MinWidth: 600,
		MaxWidth: 7690,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.Startup,
		Bind: []any{
			app,
			app.UIService,
		},
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to start application")
	}
}
