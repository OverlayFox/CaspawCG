package update

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/overlayfox/caspaw-cg/src/types"
)

type Update struct {
	logger zerolog.Logger

	casparCGClient types.CasparCGClient
	template       string
	layer          int
	videoChannels  []int

	casparMaps map[string]*Resolver // map[casparKey]*Resolver

	updateInterval time.Duration

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewUpdate(upstreamCtx context.Context, logger zerolog.Logger, template string, layer int, videoChannels []int, casparCGClient types.CasparCGClient, casparMaps map[string]*Resolver, updateInterval time.Duration) types.UpdateJob {
	ctx, cancel := context.WithCancel(upstreamCtx)
	return &Update{
		logger: logger.With().Str("component", "update").Str("template", template).Logger(),

		template:      template,
		layer:         layer,
		videoChannels: videoChannels,

		casparCGClient: casparCGClient,
		casparMaps:     casparMaps,

		updateInterval: updateInterval,

		ctx:    ctx,
		cancel: cancel,
	}
}

func (u *Update) Start() error {
	u.wg.Go(func() {
		for {
			select {
			case <-u.ctx.Done():
				return
			case <-time.After(u.updateInterval):
				casparData := make(map[string]any)
				for casparKey, resolver := range u.casparMaps {
					value, err := resolver.GetData()
					if err != nil {
						u.logger.Error().Err(err).Str("casparKey", casparKey).Msg("Failed to get data from datasource")
					}
					casparData[casparKey] = value
					resolver.Advance()
				}
				u.logger.Debug().Msgf("Updating CG data for template '%s' on layer %d, channels %v: %v", u.template, u.layer, u.videoChannels, casparData)
				err := u.casparCGClient.UpdateCGData(u.template, u.layer, u.videoChannels, casparData)
				if err != nil {
					u.logger.Error().Err(err).Msg("Failed to update CG data")
				}
			}
		}
	})

	return nil
}

func (u *Update) Stop() {
	u.cancel()
	u.wg.Wait()
}
