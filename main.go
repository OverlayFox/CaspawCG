package main

import (
	"context"
	"embed"
	"fmt"
	"os"

	casparcg "caspaw-cg/src/casparCG"
	"caspaw-cg/src/config"
	"caspaw-cg/src/data"
	"caspaw-cg/src/data/google/sheets"
	"caspaw-cg/src/events"
	"caspaw-cg/src/types"
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

	// Setup App for Front End
	ctx := context.Background()
	eventsProcessor := events.NewProcessor(ctx, logger)
	app := ui.NewApp(logger, eventsProcessor)

	// init and populate dataSource Manager
	dataSourceManager := data.NewManager(cfg.DataSourceManager)
	for _, dataSource := range cfg.DataSourceManager.GoogleSheetDataSource {
		client := sheets.NewClient(ctx, logger, dataSource)
		dataSourceManager.AddDataSource(client)
	}

	// init casparCG clients
	var casparCGClients []types.CasparCGClient
	for _, clientCfg := range cfg.CasparCGClients {
		client := casparcg.NewClient(ctx, logger, clientCfg, eventsProcessor)
		casparCGClients = append(casparCGClients, client)
	}

	err = wails.Run(&options.App{
		Title:  "caspaw-cg",
		Width:  1920,
		Height: 1080,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.Startup,
		Bind: []any{
			app,
			ui.NewUIService(
				app,
				dataSourceManager,
				casparCGClients,
			),
		},
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to start application")
	}
}
