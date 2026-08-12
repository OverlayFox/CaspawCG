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
	"github.com/overlayfox/caspaw-cg/src/types"

	"github.com/rs/zerolog"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	gs "google.golang.org/api/sheets/v4"
)

type client struct {
	logger zerolog.Logger

	cfg            d.GoogleSheetDataSource
	eventProcessor types.EventProcessor

	dataFields []*types.Data
	mtx        sync.RWMutex

	service  *gs.Service
	location *time.Location

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type Dependencies struct {
	SpreadSheetID       string
	CredentialsFilePath string
}

func NewClient(ctx context.Context, logger zerolog.Logger, cfg d.GoogleSheetDataSource, eventProcessor types.EventProcessor) (types.DataSource, error) {
	absPath, err := filepath.Abs(cfg.CredentialsFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for service account credentials file: %w", err)
	}
	jsonKey, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read service account credentials file: %w", err)
	}

	jwtConfig, err := google.JWTConfigFromJSON(jsonKey, gs.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service account credentials: %w", err)
	}

	service, err := gs.NewService(ctx, option.WithTokenSource(jwtConfig.TokenSource(ctx)))
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Sheets service: %w", err)
	}

	clientLogger := logger.With().Str("component", fmt.Sprintf("google-sheets-client-%s", cfg.SpreadSheetID)).Logger()
	location := resolveSpreadsheetLocation(ctx, clientLogger, service, cfg.SpreadSheetID)

	ctx, cancel := context.WithCancel(ctx)
	client := &client{
		logger:         clientLogger,
		cfg:            cfg,
		eventProcessor: eventProcessor,

		dataFields: make([]*types.Data, 0),

		service:  service,
		location: location,

		ctx:    ctx,
		cancel: cancel,
	}
	client.updateDataFields() // start update cycle

	return client, nil
}

// resolveSpreadsheetLocation looks up the spreadsheet's configured time zone (e.g.
// "Europe/Berlin") so serial date numbers - which carry no time zone of their own, only
// a civil wall-clock value - can be interpreted the same way Sheets itself interprets
// them. Falls back to UTC (logging why) if the property is missing or unrecognized, so a
// lookup hiccup degrades to the old behavior instead of failing client construction.
func resolveSpreadsheetLocation(ctx context.Context, logger zerolog.Logger, service *gs.Service, spreadsheetID string) *time.Location {
	spreadsheet, err := service.Spreadsheets.Get(spreadsheetID).Fields("properties.timeZone").Context(ctx).Do()
	if err != nil {
		logger.Warn().Err(err).Msg("failed to fetch spreadsheet time zone, defaulting to UTC")
		return time.UTC
	}

	tz := spreadsheet.Properties.TimeZone
	location, err := time.LoadLocation(tz)
	if err != nil {
		logger.Warn().Err(err).Str("timeZone", tz).Msg("failed to load spreadsheet time zone, defaulting to UTC")
		return time.UTC
	}

	return location
}

func (c *client) GetName() string {
	return c.cfg.Name
}

func (c *client) Prime(locations []types.Location) error {
	result, err := c.batchFetch(locations)
	if err != nil {
		return err
	}
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.dataFields = result
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

func (c *client) RemoveAllPrimes() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.dataFields = make([]*types.Data, 0)
	return nil
}

func (c *client) Get(key string) (types.Data, error) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	for _, data := range c.dataFields {
		if data.Key == key {
			return types.Data{
				Location: types.Location{
					Key:  data.Key,
					Type: data.Type,
				},
				Value: data.Value,
			}, nil
		}
	}

	return types.Data{}, fmt.Errorf("no data found for key: '%s'", key)
}

func (c *client) Close() {
	c.logger.Info().Msg("closing google sheets client")
	c.cancel()
	c.wg.Wait()
	c.logger.Info().Msg("google sheets client closed")
}

// fetch fetches a singular datapoint from the google sheet
func (c *client) batchFetch(emptyData []types.Location) ([]*types.Data, error) {
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
		// UNFORMATTED_VALUE returns each cell's raw underlying value instead of a
		// locale/display-formatted string (e.g. a datetime cell comes back as a serial
		// day number rather than a string like "19/08/2026 15:00:00" whose layout
		// depends on the spreadsheet's locale and the cell's own number format).
		ValueRenderOption("UNFORMATTED_VALUE").
		Context(c.ctx).Do()
	if err != nil {
		return nil, err
	}

	if len(resp.ValueRanges) != len(emptyData) {
		return nil, fmt.Errorf("unexpected number of value ranges in batchGet response: got %d, want %d", len(resp.ValueRanges), len(emptyData))
	}

	// Match by response order, not by string-comparing valueRange.Range against the
	// requested key: Google echoes back a canonicalized range (e.g. dropping quotes
	// that weren't strictly required), so exact string equality can silently fail to
	// match. BatchGet guarantees ValueRanges are returned in the same order as Ranges.
	result := make([]*types.Data, 0, len(resp.ValueRanges))
	for i, valueRange := range resp.ValueRanges {
		var fetchedData any
		if len(valueRange.Values) > 0 && len(valueRange.Values[0]) > 0 {
			fetchedData = valueRange.Values[0][0]
		}
		emptyDt := emptyData[i]
		if emptyDt.Type == types.DataTypeDateTime {
			if serial, ok := fetchedData.(float64); ok {
				fetchedData = c.serialToUnix(serial)
			}
		}
		result = append(result, &types.Data{
			Location: types.Location{
				Key:  emptyDt.Key,
				Type: emptyDt.Type,
			},
			Value: fetchedData,
		})
	}

	return result, nil
}

// serialToUnix converts a Sheets serial date number (as returned by the API's
// UNFORMATTED_VALUE + default SERIAL_NUMBER date-time rendering, e.g. 46145.625 for a
// half-past-three datetime) into a Unix timestamp. The serial number is a civil
// wall-clock value with no time zone of its own, so its calendar fields (year, month,
// day, hour, ...) must be read back in the spreadsheet's configured time zone
// (c.location) rather than UTC - otherwise every converted timestamp would be off by
// the spreadsheet's UTC offset.
//
// The day-count arithmetic itself is still done against a UTC epoch, not one
// constructed directly in c.location: anchoring the epoch to 1899 in a real zone risks
// resolving to that zone's pre-standardization "Local Mean Time" offset (e.g. old
// Europe/Berlin used UTC+0:53), which would throw off every date by that stale offset.
// Doing the arithmetic in UTC and only reinterpreting the resulting wall-clock fields in
// c.location avoids that: the zone lookup then applies to the real target date.
func (c *client) serialToUnix(serial float64) int64 {
	epochUTC := time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)
	wallClock := epochUTC.Add(time.Duration(serial * 24 * float64(time.Hour)))
	return time.Date(
		wallClock.Year(), wallClock.Month(), wallClock.Day(),
		wallClock.Hour(), wallClock.Minute(), wallClock.Second(), wallClock.Nanosecond(),
		c.location,
	).Unix()
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
				locations := make([]types.Location, 0, len(c.dataFields))
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
				var changed []types.DataSourceValueUpdate
				for _, data := range c.dataFields {
					for _, updatedData := range result {
						if data.Key == updatedData.Key {
							if fmt.Sprintf("%v", data.Value) != fmt.Sprintf("%v", updatedData.Value) {
								data.Value = updatedData.Value
								changed = append(changed, types.DataSourceValueUpdate{
									LocationKey: data.Key,
									Value:       updatedData.Value,
								})
							}
							break
						}
					}
				}
				c.mtx.Unlock()

				for _, ev := range changed {
					if err := c.eventProcessor.Push(ev); err != nil {
						c.logger.Error().Err(err).Str("key", ev.LocationKey).Msg("failed to emit datasource update event")
					}
				}
			}
		}
	}()
}
