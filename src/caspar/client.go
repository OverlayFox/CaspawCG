package casparcg

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/overlayfox/casparcg-amcp-go"
	casparTypes "github.com/overlayfox/casparcg-amcp-go/types"

	"github.com/rs/zerolog"

	"github.com/overlayfox/caspaw-cg/src/types"
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
	templates, err := c.caspar.Query().TLS(new(""))
	if err != nil {
		return nil, err
	}
	return templates, nil
}

func (c *client) PushCGData(template string, layer, channel int, data map[string]any, sizing types.Sizing) error {
	c.logger.Debug().Msgf("Pushing data to template '%s' on layer %d, channel %d: %v with sizing: %+v", template, layer, channel, data, sizing)

	jsonData, err := json.Marshal(data)
	if err != nil {
		c.logger.Error().Err(err).Msgf("Failed to marshal data for template '%s'", template)
		return err
	}
	jsonStr := string(jsonData)

	params := casparTypes.CGAdd{
		Template:   template,
		PlayOnLoad: true,
		Data:       &jsonStr,
	}

	if !sizing.IsDefault() {
		info, err := c.caspar.Query().Info().Generic()
		if err != nil {
			return err
		}
		res, err := VideoModeToResolution(info[0].VideoMode)
		if err != nil {
			return err
		}
		c.logger.Debug().Msgf("Setting mixer for template '%s' on layer %d, channel %d with sizing: %+v and resolution: %+v", template, layer, channel, sizing, res)
		if err = c.caspar.Mixer().Channel(channel).Layer(layer).SetFill(sizing.GetCasparMixerParams(res)); err != nil {
			return err
		}
	}
	return c.caspar.CG().Channel(channel).Layer(layer).CGLayer(1).Add(params)
}

func (c *client) StopCGData(template string, layer, channel int) error {
	c.logger.Debug().Msgf("Stopping template '%s' on layer %d, channel %d", template, layer, channel)
	if err := c.caspar.CG().Channel(channel).Layer(layer).CGLayer(1).Stop(); err != nil {
		return err
	}

	// TODO: Once the information is available via the CasparCG-AMCP-Go library, we should wait as long as the outplay time of the template is reporting
	select {
	case <-time.After(3000 * time.Millisecond):
		c.logger.Debug().Msgf("Resetting mixer for template '%s' on layer %d, channel %d to default fill", template, layer, channel)
		return c.caspar.Mixer().Channel(channel).Layer(layer).SetFill(casparTypes.MixerParamsFill{X: 0, Y: 0, XScale: 100, YScale: 100})
	case <-c.ctx.Done():
		return c.ctx.Err()
	}
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
