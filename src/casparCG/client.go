package casparcg

import (
	"context"
	"fmt"
	"sync"
	"time"

	"caspaw-cg/src/types"

	"github.com/overlayfox/casparcg-amcp-go"
	"github.com/rs/zerolog"
)

type client struct {
	logger zerolog.Logger
	cfg    *Config

	caspar         *casparcg.Client
	eventProcessor types.EventProcessor

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewClient(ctx context.Context, logger zerolog.Logger, cfg *Config, eventProcessor types.EventProcessor) types.CasparCGClient {
	c, cancel := context.WithCancel(ctx)
	client := &client{
		logger: logger.With().Str("component", fmt.Sprintf("caspar-client-%s:%d", cfg.Host, cfg.Port)).Logger(),
		cfg:    cfg,

		caspar:         casparcg.NewClient(cfg.Host, cfg.Port),
		eventProcessor: eventProcessor,

		ctx:    c,
		cancel: cancel,
	}
	client.keepAlive()
	return client
}

func (c *client) Connect() error {
	err := c.caspar.Connect()
	if err != nil {
		return err
	}
	c.keepAlive()
	return nil
}

func (c *client) GetTemplates() ([]string, error) {
	templates, _, err := c.caspar.TLS("./")
	if err != nil {
		return nil, err
	}
	return templates, nil
}

func (c *client) keepAlive() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				var event types.CasparCGKeepAlive
				_, err := c.caspar.PING("")
				if err != nil {
					c.logger.Error().Err(err).Msg("Failed to ping CasparCG server")
					event = types.CasparCGKeepAlive{
						Host:    c.cfg.Host,
						Port:    c.cfg.Port,
						IsAlive: false,
					}
				} else {
					event = types.CasparCGKeepAlive{
						Host:    c.cfg.Host,
						Port:    c.cfg.Port,
						IsAlive: true,
					}
				}
				c.eventProcessor.Push(event)
			case <-c.ctx.Done():
				return
			}
		}
	}()
}
