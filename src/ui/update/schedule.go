package update

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/overlayfox/caspaw-cg/src/types"
	"github.com/rs/zerolog"
)

type ScheduleSheet struct {
	StartTimeRange *Resolver `json:"start_time_resolver"`
	EndTimeRange   *Resolver `json:"end_time_resolver"`

	CasparMaps CasparMaps `json:"caspar_maps"` // [casparKey]Resolver
}

type Schedule struct {
	*Update

	minEvents     int // the minimum number of rows to keep on screen even when fewer than that are still active
	ScheduleSheet ScheduleSheet

	literalData map[string]any // static fields pushed alongside the resolved schedule data on every update
	sizing      types.Sizing
	delay       time.Duration

	offsetSlice   []int
	currentOffset int
}

func NewSchedule(upstreamCtx context.Context, logger zerolog.Logger, template string, layer int, videoChannels []int, casparCGClient types.CasparCGClient, casparMapsWithTiming ScheduleSheet, literalData map[string]any, sizing types.Sizing, delay, updateInterval time.Duration) types.UpdateJob {
	ctx, cancel := context.WithCancel(upstreamCtx)
	offsetSlice := make([]int, casparMapsWithTiming.StartTimeRange.GetRowAmount())
	for i := range offsetSlice {
		offsetSlice[i] = i
	}
	return &Schedule{
		Update: &Update{
			logger: logger.With().Str("component", "schedule").Str("template", template).Logger(),

			template:      template,
			layer:         layer,
			videoChannels: videoChannels,

			casparCGClient: casparCGClient,
			casparMaps:     nil,

			updateInterval: updateInterval,

			ctx:    ctx,
			cancel: cancel,
		},
		minEvents:     casparMapsWithTiming.CasparMaps.GetElementsAmount(),
		ScheduleSheet: casparMapsWithTiming,

		literalData: literalData,
		sizing:      sizing,
		delay:       delay,

		offsetSlice: offsetSlice,
	}
}

func (s *Schedule) dropPastOffsets() {
	now := time.Now()
	removedOffsets := []int{}
	keptOffsets := s.offsetSlice[:0:0]

	for idx, startOffset := range s.offsetSlice {
		nextEventIndex := (idx + s.minEvents + 1) % len(s.offsetSlice)
		nextEventOffset := s.offsetSlice[nextEventIndex]

		nextEventEnd, err := s.ScheduleSheet.EndTimeRange.GetRowData(nextEventOffset)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to get next event end time")
			continue
		}
		nextEventEndStr, ok := nextEventEnd.(int64)
		if !ok {
			s.logger.Error().Err(errors.New("invalid type assertion")).Msgf("Failed to assert next event end time as int, got %T", nextEventEnd)
			continue
		}
		nextEventEndTime := time.Unix(nextEventEndStr, 0)

		if now.After(nextEventEndTime) {
			s.logger.Debug().Int("new_offset_slice_length", len(s.offsetSlice)-len(removedOffsets)).Msg("Removed expired event from offset slice")
			removedOffsets = append(removedOffsets, startOffset)
			continue
		}
		keptOffsets = append(keptOffsets, startOffset)
	}

	// if the number of kept offsets is less than the minimum number of events, return the last few removed offsets to keep the minimum number
	if len(keptOffsets) < s.minEvents {
		amountToAdd := s.minEvents - len(keptOffsets)
		slices.Sort(removedOffsets)
		keptOffsets = append(removedOffsets[len(removedOffsets)-amountToAdd:], keptOffsets...)
	}
	s.offsetSlice = keptOffsets
}

func resolveCasparData(casparMaps CasparMaps, offset int) (map[string]any, error) {
	casparData := make(map[string]any)
	for casparKey, resolver := range casparMaps {
		data, err := resolver.GetRowData(offset)
		if err != nil {
			return nil, err
		}
		casparData[casparKey] = data
	}

	return casparData, nil
}

func (s *Schedule) Start() error {
	s.logger.Debug().Interface("offset_slice", s.offsetSlice).Msg("Starting schedule update loop")
	s.dropPastOffsets() // initial cleanup of expired events before starting the update loop
	casparData := make(map[string]any)
	s.currentOffset = s.offsetSlice[0]

	for i := range s.minEvents {
		casparMaps, err := s.ScheduleSheet.CasparMaps.GetBySlot(i)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to get caspar maps by slot")
			continue
		}
		desiredIndex := i % len(s.offsetSlice)
		eventOffset := s.offsetSlice[desiredIndex]
		casparEventData, err := resolveCasparData(casparMaps, eventOffset)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to resolve caspar data")
			continue
		}
		casparData = mergeMaps(casparData, casparEventData)
	}

	s.logger.Debug().Interface("casparData", casparData).Msg("Initial schedule update")
	if err := s.casparCGClient.AddCGData(s.template, s.layer, s.videoChannels, casparData, s.sizing, s.delay); err != nil {
		s.logger.Error().Err(err).Msg("Failed to push initial CG data")
		return err
	}

	s.wg.Go(func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-time.After(s.updateInterval):
				if len(s.offsetSlice) <= s.minEvents {
					s.logger.Info().Str("template", s.template).Int("layer", s.layer).Interface("video_channels", s.videoChannels).Msg("Stopping schedule update loop: no more events to display")
					return // terminate the update loop
				}
				s.dropPastOffsets()

				casparData := make(map[string]any)
				for idx, startOffset := range s.offsetSlice {
					if startOffset == s.currentOffset {
						nextIndex := (idx + 1) % len(s.offsetSlice)
						offset := s.offsetSlice[nextIndex] // get the actual next offset value
						s.currentOffset = offset
						for i := range s.minEvents {
							casparMaps, err := s.ScheduleSheet.CasparMaps.GetBySlot(i)
							if err != nil {
								s.logger.Error().Err(err).Msg("Failed to get caspar maps by slot")
								continue
							}

							desiredIndex := (nextIndex + i) % len(s.offsetSlice)
							eventOffset := s.offsetSlice[desiredIndex]
							casparEventData, err := resolveCasparData(casparMaps, eventOffset)
							if err != nil {
								s.logger.Error().Err(err).Msg("Failed to resolve caspar data")
								continue
							}
							casparData = mergeMaps(casparData, casparEventData)
						}
						break // only process the first matching offset to avoid duplicate updates
					}
				}

				s.logger.Debug().Interface("casparData", casparData).Msg("Schedule update")
				if err := s.casparCGClient.UpdateCGData(s.template, s.layer, s.videoChannels, casparData); err != nil {
					s.logger.Error().Err(err).Msg("Failed to update CG data")
				}
			}
		}
	})
	return nil
}
