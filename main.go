package main

import (
	"context"
	"embed"
	"os"

	"github.com/rs/zerolog"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	casparcg "github.com/overlayfox/caspaw-cg/src/caspar"
	"github.com/overlayfox/caspaw-cg/src/config"
	"github.com/overlayfox/caspaw-cg/src/data"
	"github.com/overlayfox/caspaw-cg/src/data/google/sheets"
	"github.com/overlayfox/caspaw-cg/src/events"
	"github.com/overlayfox/caspaw-cg/src/types"
	"github.com/overlayfox/caspaw-cg/src/ui"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	cfg, err := config.LoadConfig(logger.With().Str("component", "config").Logger())
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load config")
	}
	logger.Info().Msg("Config loaded successfully")

	// Setup App for Front End
	ctx := context.Background()
	eventsProcessor := events.NewProcessor(ctx, logger)
	app := ui.NewApp(ctx, logger, eventsProcessor)

	// init and populate dataSource Manager
	dataSourceManager := data.NewManager(cfg.DataSourceManager)
	if cfg.DataSourceManager != nil && cfg.DataSourceManager.GoogleSheetDataSource != nil {
		for _, dataSource := range cfg.DataSourceManager.GoogleSheetDataSource {
			client, err := sheets.NewClient(ctx, logger, dataSource, eventsProcessor)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to create Google Sheets client")
				continue
			}
			dataSourceManager.AddDataSource(client)
		}
	}

	// init casparCG clients
	casparCGClients := make([]types.CasparCGClient, 0, len(cfg.CasparCGClients))
	for _, clientCfg := range cfg.CasparCGClients {
		client := casparcg.NewClient(ctx, logger, clientCfg, eventsProcessor)
		casparCGClients = append(casparCGClients, client)
		err := client.Connect()
		if err != nil {
			logger.Error().Err(err).Str("host", clientCfg.Host).Int("port", clientCfg.Port).Msg("Failed to connect to CasparCG server")
		} else {
			logger.Debug().Str("host", clientCfg.Host).Int("port", clientCfg.Port).Msg("Connected to CasparCG server")
		}
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
