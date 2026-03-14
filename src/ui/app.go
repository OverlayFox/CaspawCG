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

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewApp creates a new App application struct
func NewApp(logger zerolog.Logger, eventProcessor types.EventProcessor) *App {
	return &App{
		logger:         logger.With().Str("component", "app").Logger(),
		eventProcessor: eventProcessor,
	}
}

func (a *App) Startup(ctx context.Context) {
	c, cancel := context.WithCancel(ctx)
	a.ctx = c
	a.cancel = cancel

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

				runtime.EventsEmit(a.ctx, string(identifier), payload)
			}
		}
	}()
}
