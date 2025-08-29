package updater

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/OverlayFox/CaspawCG/types"
	"github.com/rs/zerolog"
)

type Updater struct {
	logger zerolog.Logger

	sheetsHandler types.SheetsHandler

	updateCh chan string

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewHandler(updateCh chan string, sheetsHandler types.SheetsHandler, logger zerolog.Logger) types.Updater {
	ctx, cancel := context.WithCancel(context.Background())
	return &Updater{
		ctx:           ctx,
		cancel:        cancel,
		updateCh:      updateCh,
		sheetsHandler: sheetsHandler,
		logger:        logger,
	}
}

func (u *Updater) Start(layer int) {
	u.logger.Info().Msg("starting updater")

	u.wg.Add(1)
	go func() {
		defer u.wg.Done()

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		startIndex := 1

		for {
			select {
			case <-u.ctx.Done():
				return
			case <-ticker.C:
				jsonData, err := u.sheetsHandler.GetCurrentSchedule(startIndex)
				if err != nil {
					u.logger.Error().Err(err).Msg("failed to get current schedule")
					continue
				}
				startIndex++

				cmd := []string{
					"CG",
					fmt.Sprintf("1-%d", layer),
					"UPDATE",
					"1",
					fmt.Sprintf("\"%s\"", string(jsonData)),
				}

				select {
				case u.updateCh <- strings.Join(cmd, " "):
					u.logger.Debug().Msg("sent schedule update")
				case <-u.ctx.Done():
					return
				default:
					// Do nothing
				}
			}
		}
	}()
}

func (u *Updater) Stop() {
	u.logger.Info().Msg("Stopping updater")

	u.cancel()
	u.wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	u.ctx = ctx
	u.cancel = cancel
}

func (u *Updater) Close() {
	u.logger.Info().Msg("Closing updater")

	u.cancel()
	u.wg.Wait()
}
