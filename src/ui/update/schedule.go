package update

import (
	"context"
	"time"

	"github.com/overlayfox/caspaw-cg/src/types"
	"github.com/rs/zerolog"
)

type Schedule struct {
	*Update

	currentAbsRowNumber int // the absolute row number the current displayed schedule is on
	minElements         int // the minimum number of elements to display in the schedule
}

func NewSchedule(upstreamCtx context.Context, logger zerolog.Logger, template string, layer int, videoChannels []int, casparCGClient types.CasparCGClient, casparMaps map[string]*Resolver, updateInterval time.Duration, minElements int, startTimeColumn, endTimeColumn string) types.UpdateJob {
	ctx, cancel := context.WithCancel(upstreamCtx)
	return &Schedule{
		Update: &Update{
			logger: logger.With().Str("component", "schedule").Str("template", template).Logger(),

			template:      template,
			layer:         layer,
			videoChannels: videoChannels,

			casparCGClient: casparCGClient,
			casparMaps:     casparMaps,

			updateInterval: updateInterval,

			ctx:    ctx,
			cancel: cancel,
		},
		currentAbsRowNumber: 0,
		minElements:         minElements,
	}
}

func (s *Schedule) Start() error {
	s.wg.Go(func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-time.After(s.updateInterval):
				casparData := make(map[string]any)
				for casparKey, resolver := range s.casparMaps {
					values, err := resolver.GetAllData()
					if err != nil {
						s.logger.Error().Err(err).Str("casparKey", casparKey).Msg("Failed to get data from datasource")
						continue
					}
					casparData[casparKey] = values
				}
				s.logger.Info().Any("casparData", casparData).Msg("Fetched schedule data")
			}
		}
	})
	return nil
}
