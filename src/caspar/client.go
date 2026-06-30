package casparcg

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/overlayfox/casparcg-amcp-go"
	casparTypes "github.com/overlayfox/casparcg-amcp-go/types"
	"github.com/overlayfox/casparcg-amcp-go/types/responses"

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
	return client
}

func (c *client) Connect() error {
	defer c.keepAlive()

	err := c.caspar.Connect(c.ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *client) GetTemplates() ([]string, error) {
	return c.caspar.Query().TLS(new(""))
}

func (c *client) GetMedia() ([]string, error) {
	media, err := c.caspar.Query().CLS(new(""))
	if err != nil {
		return nil, err
	}
	result := make([]string, len(media))
	for i, m := range media {
		result[i] = m.Filename
	}
	return result, nil
}

func (c *client) GetMediaInfo(filename string) (responses.CINF, error) {
	return c.caspar.Query().CINF(filename)
}

func (c *client) AddCGData(template string, layer int, channels []int, data map[string]any, sizing types.Sizing, delay time.Duration) error {
	c.logger.Debug().Msgf("Adding data to template '%s' on layer %d, channels %v: %v with sizing: %+v and delay: %v", template, layer, channels, data, sizing, delay)

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

	info, err := c.caspar.Query().Info().Generic()
	if err != nil {
		return err
	}
	res, err := VideoModeToResolution(info[0].VideoMode)
	if err != nil {
		return err
	}
	for _, channel := range channels {
		c.logger.Debug().Msgf("Setting mixer for template '%s' on layer %d, channel %d with sizing: %+v and resolution: %+v", template, layer, channel, sizing, res)
		if err = c.caspar.Mixer().Channel(channel).Layer(layer).SetFill(sizing.GetCasparMixerParams(res)); err != nil {
			return err
		}
	}

	if delay > 0 {
		select {
		case <-time.After(delay):
			// continue to play the CG data after the delay
		case <-c.ctx.Done():
			return c.ctx.Err()
		}
	}

	for _, channel := range channels {
		if err := c.caspar.CG().Channel(channel).Layer(layer).CGLayer(1).Add(params); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) StopCGData(template string, layer int, channels []int, delay time.Duration) error {
	c.logger.Debug().Msgf("Stopping template '%s' on layer %d, channels %v with delay: %v", template, layer, channels, delay)

	if delay > 0 {
		select {
		case <-time.After(delay):
			// continue to stop the CG data after the delay
		case <-c.ctx.Done():
			return c.ctx.Err()
		}
	}

	for _, channel := range channels {
		if err := c.caspar.CG().Channel(channel).Layer(layer).CGLayer(1).Stop(); err != nil {
			return err
		}
	}
	return nil
	// TODO: Once the information is available via the CasparCG-AMCP-Go library, we should wait as long as the outplay time of the template is reporting
	// select {
	// case <-time.After(6000 * time.Millisecond):
	// 	c.logger.Debug().Msgf("Resetting mixer for template '%s' on layer %d, channel %d to default fill", template, layer, channel)
	// 	return c.caspar.Mixer().Channel(channel).Layer(layer).SetFill(casparTypes.MixerParamsFill{X: 0, Y: 0, XScale: 1, YScale: 1})
	// case <-c.ctx.Done():
	// 	return c.ctx.Err()
	// }
}

func (c *client) NextCGData(template string, layer int, channels []int, delay time.Duration) error {
	c.logger.Debug().Msgf("Nexting template '%s' on layer %d, channels %v with delay: %v", template, layer, channels, delay)

	if delay > 0 {
		select {
		case <-time.After(delay):
			// continue to next the CG data after the delay
		case <-c.ctx.Done():
			return c.ctx.Err()
		}
	}

	for _, channel := range channels {
		if err := c.caspar.CG().Channel(channel).Layer(layer).CGLayer(1).Next(); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) UpdateCGData(template string, layer int, channels []int, data map[string]any) error {
	c.logger.Debug().Msgf("Updating data for template '%s' on layer %d, channels %v: %v", template, layer, channels, data)

	jsonData, err := json.Marshal(data)
	if err != nil {
		c.logger.Error().Err(err).Msgf("Failed to marshal data for template '%s'", template)
		return err
	}
	jsonStr := string(jsonData)

	for _, channel := range channels {
		if err := c.caspar.CG().Channel(channel).Layer(layer).CGLayer(1).Update(jsonStr); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) PlayMedia(filename string, layer int, channels []int, loop bool, delay time.Duration) error {
	c.logger.Debug().Msgf("Playing media '%s' on layer %d, channels %v (loop=%v) with delay: %v", filename, layer, channels, loop, delay)

	if delay > 0 {
		select {
		case <-time.After(delay):
		case <-c.ctx.Done():
			return c.ctx.Err()
		}
	}

	params := casparTypes.LayerPlay{ClipName: &filename}
	if loop {
		loopParams := []string{"LOOP"}
		params.Parameters = &loopParams
	}
	for _, channel := range channels {
		if err := c.caspar.Layer().Channel(channel).Layer(layer).Play(params); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) StopMedia(layer int, channels []int, delay time.Duration) error {
	c.logger.Debug().Msgf("Stopping media on layer %d, channels %v with delay: %v", layer, channels, delay)

	if delay > 0 {
		select {
		case <-time.After(delay):
		case <-c.ctx.Done():
			return c.ctx.Err()
		}
	}

	for _, channel := range channels {
		if err := c.caspar.Layer().Channel(channel).Layer(layer).Stop(); err != nil {
			return err
		}
	}
	return nil
}

func (c *client) ClearChannels(channels []int) {
	c.logger.Debug().Msgf("Clearing CG data on channels: %v", channels)
	for _, channel := range channels {
		c.caspar.Layer().Channel(channel).Clear()
		c.caspar.Mixer().Channel(channel).Clear()
	}
}

// TODO: this is horrible and needs to be rethought, we shouldn't need to loop through all channels and layers to clear them....but it will do for now as a emergency stop.
func (c *client) ClearAll() {
	c.logger.Debug().Msgf("Clearing all CG data on channel all channels")
	for i := range 9999 {
		c.caspar.Layer().Channel(i).Clear()
		c.caspar.Mixer().Channel(i).Clear()
	}
}

func (c *client) keepAlive() {
	c.wg.Go(func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		sentDebugMessage := false
		for {
			select {
			case <-ticker.C:
				var event types.CasparCGKeepAlive
				_, err := c.caspar.Ping(nil)
				if err != nil {
					if !sentDebugMessage {
						c.logger.Debug().Err(err).Msg("Failed to ping CasparCG server")
						sentDebugMessage = true
					}
					event = types.CasparCGKeepAlive{
						Host:    c.cfg.Host,
						Port:    c.cfg.Port,
						IsAlive: false,
					}
					err := c.caspar.Connect(c.ctx)
					if err == nil {
						c.logger.Debug().Msg("Reconnected to CasparCG server")
					}
				} else {
					sentDebugMessage = false
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
	})
}

func (c *client) Close() {
	c.cancel()
	c.wg.Wait()
	c.caspar.Close()
}
