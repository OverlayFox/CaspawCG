package events

import (
	"context"
	"sync"

	"caspaw-cg/src/types"

	"github.com/rs/zerolog"
)

type processor struct {
	logger zerolog.Logger

	channels []chan types.Event
	mtx      sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
}

func NewProcessor(ctx context.Context, logger zerolog.Logger) types.EventProcessor {
	c, cancel := context.WithCancel(ctx)
	return &processor{
		logger: logger.With().Str("component", "events_processor").Logger(),
		ctx:    c,
		cancel: cancel,
	}
}

func (p *processor) Push(event types.Event) error {
	p.logger.Debug().Interface("event", event).Msg("received event, distributing it to all listeners")

	p.mtx.Lock()
	defer p.mtx.Unlock()

	for _, ch := range p.channels {
		select {
		case ch <- event:
		case <-p.ctx.Done():
			return p.ctx.Err()
		}
	}

	return nil
}

func (p *processor) Listen() <-chan types.Event {
	ch := make(chan types.Event)

	p.mtx.Lock()
	p.channels = append(p.channels, ch)
	p.mtx.Unlock()

	return ch
}

func (p *processor) CloseChannel(ch <-chan types.Event) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	for i, c := range p.channels {
		if c == ch {
			close(c)
			p.channels = append(p.channels[:i], p.channels[i+1:]...)
			break
		}
	}
}

func (p *processor) Cancel() {
	p.cancel()

	p.mtx.Lock()
	for _, ch := range p.channels {
		close(ch)
	}
	p.mtx.Unlock()
}
