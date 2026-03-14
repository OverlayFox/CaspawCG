package sheets

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"caspaw-cg/src/types"

	"github.com/rs/zerolog"
	"google.golang.org/api/option"
	gs "google.golang.org/api/sheets/v4"
)

type client struct {
	logger zerolog.Logger
	deps   Dependencies

	dataFields map[string]any
	mtx        sync.RWMutex

	service *gs.Service

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type Dependencies struct {
	SpreadSheetID       string
	CredentialsFilePath string
}

func NewClient(ctx context.Context, logger zerolog.Logger, deps Dependencies) types.DataSource {
	absPath, err := filepath.Abs(deps.CredentialsFilePath)
	if err != nil {
		logger.Error().Err(err).Msg("invalid credentials file path")
		return nil
	}
	if _, err := os.Stat(absPath); err != nil {
		logger.Error().Err(err).Msg("credentials file does not exist or is inaccessible")
		return nil
	}

	service, err := gs.NewService(ctx, option.WithAuthCredentialsFile(option.ServiceAccount, deps.CredentialsFilePath))
	if err != nil {
		logger.Error().Err(err).Msg("failed to create Google Sheets service")
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	client := &client{
		logger: logger,
		deps:   deps,

		service: service,

		ctx:    ctx,
		cancel: cancel,
	}
	client.updateDataFields() // start update cycle

	return client
}

func (c *client) Prime(locations []string) error {
	result, err := c.batchFetch(locations)
	if err != nil {
		return err
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()
	maps.Copy(c.dataFields, result)

	return nil
}

func (c *client) RemovePrime(locations []string) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	for _, loc := range locations {
		delete(c.dataFields, loc)
	}
	return nil
}

func (c *client) Get(location string) (any, error) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	value, ok := c.dataFields[location]
	if !ok {
		return nil, fmt.Errorf("no data found at location: '%s'", location)
	}
	return value, nil
}

func (c *client) Close() {
	c.logger.Info().Msg("closing google sheets client")
	c.cancel()
	c.wg.Wait()
	c.logger.Info().Msg("google sheets client closed")
}

// fetch fetches a singular datapoint from the google sheet
func (c *client) batchFetch(locations []string) (map[string]any, error) {
	if len(locations) == 0 {
		return nil, fmt.Errorf("no locations provided")
	}
	for _, loc := range locations {
		if strings.Contains(loc, ":") || !strings.Contains(loc, "!") {
			return nil, fmt.Errorf("invalid location format '%s', use 'sheet1!A1'", loc)
		}
	}

	resp, err := c.service.Spreadsheets.Values.
		BatchGet(c.deps.SpreadSheetID).
		Ranges(locations...).
		Context(c.ctx).Do()
	if err != nil {
		return nil, err
	}

	result := make(map[string]any)
	for i, valueRange := range resp.ValueRanges {
		loc := locations[i]
		if len(valueRange.Values) > 0 && len(valueRange.Values[0]) > 0 {
			result[loc] = valueRange.Values[0][0]
		} else {
			c.logger.Warn().Str("location", loc).Msg("cell is empty, adding it to result with nil value")
			result[loc] = nil
		}
	}
	return result, nil
}

func (c *client) updateDataFields() {
	c.wg.Add(1)
	ticker := time.NewTicker(10 * time.Second) // TODO: make this configurable via UI

	go func() {
		defer c.wg.Done()
		defer ticker.Stop()

		for {
			select {
			case <-c.ctx.Done():
				return
			case <-ticker.C:
				c.mtx.RLock()
				locations := make([]string, 0, len(c.dataFields))
				for loc := range c.dataFields {
					locations = append(locations, loc)
				}
				c.mtx.RUnlock()

				if len(locations) == 0 {
					continue
				}

				result, err := c.batchFetch(locations)
				if err != nil {
					c.logger.Error().Err(err).Msg("failed to fetch data fields")
					continue
				}

				c.mtx.Lock()
				maps.Copy(c.dataFields, result)
				c.mtx.Unlock()
			}
		}
	}()
}
