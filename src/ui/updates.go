package ui

import (
	"context"
	"errors"
	"sync"
	"time"

	guuid "github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/overlayfox/caspaw-cg/src/types"
)

type Resolver struct {
	offset     int
	datasource types.DataSource
	dataRange  types.Range
}

func NewResolver(datasource types.DataSource, dataRange types.Range, offset int) Resolver {
	return Resolver{
		datasource: datasource,
		dataRange:  dataRange,
		offset:     offset,
	}
}

func (r *Resolver) GetData() (any, error) {
	if r.offset < 0 || r.offset >= len(r.dataRange.Locations) {
		return nil, errors.New("offset out of range")
	}
	key := r.dataRange.Locations[r.offset].Key
	data, err := r.datasource.Get(key)
	if err != nil {
		return nil, err
	}
	return data.Value, nil
}

type Update struct {
	logger zerolog.Logger

	casparCGClient types.CasparCGClient
	template       string
	layer          int
	videoChannels  []int

	casparMaps map[string]Resolver // map[casparKey]Resolver

	updateInterval time.Duration

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewUpdate(upstreamCtx context.Context, logger zerolog.Logger, template string, layer int, videoChannels []int, casparCGClient types.CasparCGClient, casparMaps map[string]Resolver, updateInterval time.Duration) types.UpdateJob {
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
					data, err := resolver.datasource.Get(resolver.dataRange.Locations[resolver.offset].Key)
					if err != nil {
						u.logger.Error().Err(err).Str("key", resolver.dataRange.Locations[resolver.offset].Key).Msg("Failed to get data from datasource")
					}
					casparData[casparKey] = data.Value

					resolver.offset++
					if resolver.offset >= len(resolver.dataRange.Locations) {
						resolver.offset = 0 // Reset to the beginning if we reach the end
					}
				}

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

type UpdateHandler struct {
	logger zerolog.Logger

	datasourceManager types.DatasourceManager
	casparCGClients   types.CasparCGClient

	cycles map[string]types.UpdateJob

	ctx    context.Context
	cancel context.CancelFunc
}

func NewUpdateHandler(upstreamCtx context.Context, logger zerolog.Logger, datasourceManager types.DatasourceManager, casparCGClients types.CasparCGClient) *UpdateHandler {
	ctx, cancel := context.WithCancel(upstreamCtx)
	return &UpdateHandler{
		logger: logger.With().Str("component", "update-handler").Logger(),

		datasourceManager: datasourceManager,
		casparCGClients:   casparCGClients,

		cycles: make(map[string]types.UpdateJob),

		ctx:    ctx,
		cancel: cancel,
	}
}

func (u *UpdateHandler) AddUpdateJob(template string, layer int, videoChannels []int, casparCGClient types.CasparCGClient, casparMaps map[string]Resolver, updateInterval time.Duration) (uuid string) {
	uuid = guuid.NewString()
	u.logger.Debug().Str("uuid", uuid).Msg("Adding update job")

	job := NewUpdate(u.ctx, u.logger, template, layer, videoChannels, casparCGClient, casparMaps, updateInterval)
	u.cycles[uuid] = job
	job.Start()

	return uuid
}

func (u *UpdateHandler) RemoveUpdateJob(uuid string) error {
	u.logger.Debug().Str("uuid", uuid).Msg("Removing update job")

	if job, ok := u.cycles[uuid]; ok {
		job.Stop()
		delete(u.cycles, uuid)
		return nil
	}
	return errors.New("update job not found")
}
