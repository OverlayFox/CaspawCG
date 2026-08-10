package update

import (
	"context"
	"errors"
	"time"

	guuid "github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/overlayfox/caspaw-cg/src/types"
)

type Handler struct {
	logger zerolog.Logger

	datasourceManager types.DatasourceManager
	casparCGClients   types.CasparCGClient

	cycles map[string]types.UpdateJob

	ctx    context.Context
	cancel context.CancelFunc
}

func NewUpdateHandler(upstreamCtx context.Context, logger zerolog.Logger, datasourceManager types.DatasourceManager, casparCGClients types.CasparCGClient) *Handler {
	ctx, cancel := context.WithCancel(upstreamCtx)
	return &Handler{
		logger: logger.With().Str("component", "update-handler").Logger(),

		datasourceManager: datasourceManager,
		casparCGClients:   casparCGClients,

		cycles: make(map[string]types.UpdateJob),

		ctx:    ctx,
		cancel: cancel,
	}
}

func (u *Handler) AddUpdateJob(template string, layer int, videoChannels []int, casparCGClient types.CasparCGClient, casparMaps map[string]*Resolver, updateInterval time.Duration) (uuid string) {
	uuid = guuid.NewString()
	u.logger.Debug().Str("uuid", uuid).Msg("Adding update job")

	job := NewUpdate(u.ctx, u.logger, template, layer, videoChannels, casparCGClient, casparMaps, updateInterval)
	u.cycles[uuid] = job
	job.Start()

	return uuid
}

func (u *Handler) AddScheduleJob(template string, layer int, videoChannels []int, casparCGClient types.CasparCGClient, casparMaps map[string]*Resolver, updateInterval time.Duration, startTimeColumn, endTimeColumn string) (uuid string) {
	uuid = guuid.NewString()
	u.logger.Debug().Str("uuid", uuid).Msg("Adding schedule job")

	job := NewSchedule(u.ctx, u.logger, template, layer, videoChannels, casparCGClient, casparMaps, updateInterval, startTimeColumn, endTimeColumn)
	u.cycles[uuid] = job
	job.Start()

	return uuid
}

func (u *Handler) RemoveUpdateJob(uuid string) error {
	u.logger.Debug().Str("uuid", uuid).Msg("Removing update job")

	if job, ok := u.cycles[uuid]; ok {
		job.Stop()
		delete(u.cycles, uuid)
		return nil
	}
	return errors.New("update job not found")
}
