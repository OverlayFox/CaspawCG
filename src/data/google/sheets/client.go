package sheets

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	d "github.com/overlayfox/caspaw-cg/src/data"

	"github.com/rs/zerolog"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	gs "google.golang.org/api/sheets/v4"
)

type client struct {
	logger zerolog.Logger

	cfg d.GoogleSheetDataSource

	dataFields []*d.Data
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

func NewClient(ctx context.Context, logger zerolog.Logger, cfg d.GoogleSheetDataSource) d.DataSource {
	absPath, err := filepath.Abs(cfg.CredentialsFilePath)
	if err != nil {
		logger.Error().Err(err).Msg("invalid credentials file path")
		return nil
	}
	jsonKey, err := os.ReadFile(absPath)
	if err != nil {
		logger.Error().Err(err).Msg("credentials file does not exist or is inaccessible")
		return nil
	}

	jwtConfig, err := google.JWTConfigFromJSON(jsonKey, gs.SpreadsheetsScope)
	if err != nil {
		logger.Error().Err(err).Msg("failed to parse service account credentials")
		return nil
	}

	service, err := gs.NewService(ctx, option.WithTokenSource(jwtConfig.TokenSource(ctx)))
	if err != nil {
		logger.Error().Err(err).Msg("failed to create Google Sheets service")
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	client := &client{
		logger: logger.With().Str("component", fmt.Sprintf("google-sheets-client-%s", cfg.SpreadSheetID)).Logger(),
		cfg:    cfg,

		dataFields: make([]*d.Data, 0),

		service: service,

		ctx:    ctx,
		cancel: cancel,
	}
	client.updateDataFields() // start update cycle

	return client
}

func (c *client) GetName() string {
	return fmt.Sprintf("GoogleSheet: %s", c.cfg.SpreadSheetID)
}

func (c *client) Prime(locations []d.Location) error {
	result, err := c.batchFetch(locations)
	if err != nil {
		return err
	}
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.dataFields = append(c.dataFields, result...)
	return nil
}

func (c *client) RemovePrime(keys []string) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	for _, key := range keys {
		for i, data := range c.dataFields {
			if data.Key == key {
				c.dataFields = append(c.dataFields[:i], c.dataFields[i+1:]...)
				break
			}
		}
	}
	return nil
}

func (c *client) Get(key string) (d.Data, error) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	for _, data := range c.dataFields {
		if data.Key == key {
			return d.Data{
				Location: d.Location{
					Key:  data.Key,
					Type: data.Type,
				},
				Value: data.Value,
			}, nil
		}
	}

	return d.Data{}, fmt.Errorf("no data found for key: '%s'", key)
}

func (c *client) Close() {
	c.logger.Info().Msg("closing google sheets client")
	c.cancel()
	c.wg.Wait()
	c.logger.Info().Msg("google sheets client closed")
}

// fetch fetches a singular datapoint from the google sheet
func (c *client) batchFetch(emptyData []d.Location) ([]*d.Data, error) {
	if len(emptyData) == 0 {
		return nil, errors.New("no locations provided")
	}
	for _, loc := range emptyData {
		if strings.Contains(loc.Key, ":") || !strings.Contains(loc.Key, "!") {
			return nil, fmt.Errorf("invalid location format '%s', use 'sheet1!A1'", loc.Key)
		}
	}

	keys := make([]string, 0, len(emptyData))
	for _, loc := range emptyData {
		keys = append(keys, loc.Key)
	}
	resp, err := c.service.Spreadsheets.Values.
		BatchGet(c.cfg.SpreadSheetID).
		Ranges(keys...).
		Context(c.ctx).Do()
	if err != nil {
		return nil, err
	}

	result := make([]*d.Data, 0, len(resp.ValueRanges))
	for _, valueRange := range resp.ValueRanges {
		var fetchedData any
		if len(valueRange.Values) > 0 && len(valueRange.Values[0]) > 0 {
			fetchedData = valueRange.Values[0][0]
		} else {
			fetchedData = nil
		}
		for _, emptyDt := range emptyData {
			if emptyDt.Key == valueRange.Range {
				result = append(result, &d.Data{
					Location: d.Location{
						Key:  emptyDt.Key,
						Type: emptyDt.Type,
					},
					Value: fetchedData,
				})
				break
			}
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
				locations := make([]d.Location, 0, len(c.dataFields))
				for _, data := range c.dataFields {
					locations = append(locations, data.Location)
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
				for _, data := range c.dataFields {
					for _, updatedData := range result {
						if data.Key == updatedData.Key {
							data.Value = updatedData.Value
							break
						}
					}
				}
				c.mtx.Unlock()
			}
		}
	}()
}
