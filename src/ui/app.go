package ui

import (
	"context"
	"sync"

	"github.com/rs/zerolog"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	casparcg "github.com/overlayfox/caspaw-cg/src/caspar"
	"github.com/overlayfox/caspaw-cg/src/config"
	"github.com/overlayfox/caspaw-cg/src/data"
	"github.com/overlayfox/caspaw-cg/src/data/google/sheets"
	"github.com/overlayfox/caspaw-cg/src/events"
	"github.com/overlayfox/caspaw-cg/src/types"
)

// App struct
type App struct {
	logger zerolog.Logger

	UIService         *UIService
	dataSourceManager types.DatasourceManager
	casparCGClient    types.CasparCGClient
	eventProcessor    types.EventProcessor

	wailsCtx context.Context // opaque key for identifying with Wails runtime

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewApp creates a new App application struct
func NewApp(upstreamCtx context.Context, logger zerolog.Logger, config *config.Config) (*App, error) {
	ctx, cancel := context.WithCancel(upstreamCtx)

	eventsProcessor := events.NewProcessor(ctx, logger)

	datasourceManager := data.NewManager(config.DataSourceManager)
	if config.DataSourceManager != nil && config.DataSourceManager.GoogleSheetDataSource != nil {
		for _, dataSource := range config.DataSourceManager.GoogleSheetDataSource {
			client, err := sheets.NewClient(ctx, logger, dataSource, eventsProcessor)
			if err != nil {
				cancel()
				return nil, err
			}
			datasourceManager.AddDataSource(client)
		}
	}

	casparClient := casparcg.NewClient(ctx, logger, config.CasparCGClient, eventsProcessor)
	err := casparClient.Connect()
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to connect to CasparCG server")
	} else {
		logger.Debug().Str("host", config.CasparCGClient.Host).Int("port", config.CasparCGClient.Port).Msg("Connected to CasparCG server")
	}

	a := &App{
		logger: logger.With().Str("component", "app").Logger(),

		eventProcessor:    eventsProcessor,
		dataSourceManager: datasourceManager,
		casparCGClient:    casparClient,

		ctx:    ctx,
		cancel: cancel,
	}
	a.UIService = NewUIService(ctx, a, datasourceManager, casparClient)

	return a, nil
}

func (a *App) Startup(wailsCtx context.Context) {
	a.wailsCtx = wailsCtx
	a.listenForEvents()
}

func (a *App) Shutdown() {
	a.cancel()

	a.eventProcessor.Close()
	a.dataSourceManager.Close()
	a.casparCGClient.Close()
	a.UIService.Close()

	a.wg.Wait()
}

// listenForEvents continuously listens for events from the event processor
// and emits them to the frontend
func (a *App) listenForEvents() {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		eventCh := a.eventProcessor.Listen()
		for {
			select {
			case <-a.ctx.Done():
				return
			case event := <-eventCh:
				identifier := event.GetIdentifier()
				data := event.GetData()

				payload := types.WailsPayload{
					Identifier: string(identifier),
					Value:      data,
				}
				runtime.EventsEmit(a.wailsCtx, "live-data-update", payload)
			}
		}
	}()
}
