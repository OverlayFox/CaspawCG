package ui

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/overlayfox/caspaw-cg/src/types"
)

type Resolver struct {
	datasource types.DataSource
	dataRange  types.Range
}

func NewResolver(datasource types.DataSource, dataRange types.Range) Resolver {
	return Resolver{
		datasource: datasource,
		dataRange:  dataRange,
	}
}

type Update struct {
	logger zerolog.Logger

	template      string
	layer         int
	videoChannels []int

	casarCGClient types.CasparCGClient
	casparMaps    map[string]Resolver

	offset int
	delay  time.Duration

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewUpdate(upstreamCtx context.Context, logger zerolog.Logger, template string, layer int, videoChannels []int, casparCGClient types.CasparCGClient, casparMaps map[string]Resolver, offset int, delay time.Duration) *Update {
	ctx, cancel := context.WithCancel(upstreamCtx)
	return &Update{
		logger: logger.With().Str("component", "update").Str("template", template).Logger(),

		template:      template,
		layer:         layer,
		videoChannels: videoChannels,

		casarCGClient: casparCGClient,
		casparMaps:    casparMaps,

		offset: offset,
		delay:  delay,

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
			case <-time.After(u.delay):
				casparData := make(map[string]any)
				for casparKey, resolver := range u.casparMaps {
					data, err := resolver.datasource.Get(resolver.dataRange.Locations[u.offset].Key)
					if err != nil {
						u.logger.Error().Err(err).Str("key", resolver.dataRange.Locations[u.offset].Key).Msg("Failed to get data from datasource")
					}
					casparData[casparKey] = data.Value
				}

				err := u.casarCGClient.UpdateCGData(u.template, u.layer, u.videoChannels, casparData)
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
	casparCGClients   []types.CasparCGClient

	cycles map[string]types.UpdateJob

	ctx    context.Context
	cancel context.CancelFunc
}

func NewUpdateHandler(upstreamCtx context.Context, logger zerolog.Logger, datasourceManager types.DatasourceManager, casparCGClients []types.CasparCGClient) *UpdateHandler {
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

func (u *UpdateHandler) AddUpdateJob(uuid string, job types.UpdateJob) {
	u.logger.Debug().Str("uuid", uuid).Msg("Adding update job")

	u.cycles[uuid] = job
	job.Start()
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
