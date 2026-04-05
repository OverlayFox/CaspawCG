package casparcg

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"caspaw-cg/src/types"

	"github.com/overlayfox/casparcg-amcp-go"
	casparTypes "github.com/overlayfox/casparcg-amcp-go/types"
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
	err := c.caspar.Connect(c.ctx)
	if err != nil {
		return err
	}
	c.keepAlive()
	return nil
}

func (c *client) GetTemplates() ([]string, error) {
	templates, err := c.caspar.Query().TLS(ptr(""))
	if err != nil {
		return nil, err
	}
	return templates, nil
}

func (c *client) PushCGData(template string, layer, channel int, data map[string]any, sizing types.Sizing) error {
	c.logger.Info().Msgf("Pushing data to template '%s' on layer %d, channel %d: %v with sizing: %+v", template, layer, channel, data, sizing)

	_, err := c.caspar.Query().Info().Template(template) // Using this to check if the template exists before trying to add data to it
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		c.logger.Error().Err(err).Msgf("Failed to marshal data for template '%s'", template)
		return err
	}
	jsonStr := string(jsonData)

	params := casparTypes.CGAdd{
		Template:   template,
		PlayOnLoad: true,
		Data:       ptr(jsonStr),
	}
	return c.caspar.CG().Channel(channel).Layer(layer).CGLayer(1).Add(params)
}

func (c *client) StopCGData(template string, layer, channel int) error {
	return c.caspar.CG().Channel(channel).Layer(layer).CGLayer(1).Stop()
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
				_, err := c.caspar.Ping(nil)
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

func ptr[T any](t T) *T {
	return &t
}
