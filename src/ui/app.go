package ui

import (
	"context"
	"sync"

	"caspaw-cg/src/types"

	"github.com/rs/zerolog"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	logger zerolog.Logger

	eventProcessor types.EventProcessor

	ctx      context.Context
	wailsCtx context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewApp creates a new App application struct
func NewApp(ctx context.Context, logger zerolog.Logger, eventProcessor types.EventProcessor) *App {
	c, cancel := context.WithCancel(ctx)
	return &App{
		logger:         logger.With().Str("component", "app").Logger(),
		eventProcessor: eventProcessor,
		ctx:            c,
		cancel:         cancel,
	}
}

func (a *App) Startup(wailsCtx context.Context) {
	a.wailsCtx = wailsCtx
	a.listenForEvents()
}

func (a *App) Shutdown() {
	a.cancel()
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
				a.logger.Info().Interface("payload", payload).Msg("Received event")

				runtime.EventsEmit(a.wailsCtx, "live-data-update", payload)
			}
		}
	}()
}
