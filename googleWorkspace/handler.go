package googleworkspace

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/OverlayFox/CaspawCG/types"
	"github.com/rs/zerolog"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Handler struct {
	logger zerolog.Logger

	sheets *sheets.SpreadsheetsService
	cgID   string

	countdown           *types.SimpleCountdown
	countdownDuration   *types.SimpleCountdown
	djCountdown         *types.DJCountdown
	djCountdownDuration *types.DJCountdown
	lowerThird1         *types.LowerThird
	lowerThird2         *types.LowerThird
	lowerThirdDJ        *types.LowerThird
	schedule            []*types.ScheduleRow

	mtx         sync.RWMutex
	scheduleMtx sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewHandler(logger zerolog.Logger) (types.SheetsData, error) {
	ctx, cancel := context.WithCancel(context.Background())
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("service-account.json"))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create Sheets service: %w", err)
	}

	return &Handler{
		logger: logger,

		sheets: srv.Spreadsheets,
		cgID:   "1FneiBwBkRFOWzGvuMkOuJgV3aaBDaDNS5Lp-IsCsugA",
		ctx:    ctx,
		cancel: cancel,

		countdownDuration: &types.SimpleCountdown{
			CountdownType: types.CountdownTypeDuration,
			Text:          "Countdown",
			CountdownTime: "00:00:00",
		},
		countdown: &types.SimpleCountdown{
			CountdownType: types.CountdownTypeToTime,
			Text:          "Countdown",
			CountdownTime: "00:00:00",
		},
		djCountdownDuration: &types.DJCountdown{
			SimpleCountdown: &types.SimpleCountdown{
				CountdownType: types.CountdownTypeDuration,
				Text:          "DJ Countdown",
				CountdownTime: "00:00:00",
			},
			FirstDJ:  "<First DJ>",
			SecondDJ: "<Second DJ>",
		},
		djCountdown: &types.DJCountdown{
			SimpleCountdown: &types.SimpleCountdown{
				CountdownType: types.CountdownTypeToTime,
				Text:          "DJ Countdown",
				CountdownTime: "00:00:00",
			},
			FirstDJ:  "<First DJ>",
			SecondDJ: "<Second DJ>",
		},
		lowerThird1: &types.LowerThird{
			Row1: "",
			Row2: "",
		},
		lowerThird2: &types.LowerThird{
			Row1: "",
			Row2: "",
		},
		schedule: make([]*types.ScheduleRow, 0),
	}, nil
}

func (h *Handler) Start() {
	h.logger.Info().Msg("Starting Google Sheets handler")

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()

		h.logger.Debug().Msg("Pulling initial data from Google Sheets")
		h.pullCgSheet()

		ticker5sec := time.NewTicker(5 * time.Second)
		defer ticker5sec.Stop()

		for {
			select {
			case <-ticker5sec.C:
				h.pullCgSheet()
			case <-h.ctx.Done():
				h.logger.Debug().Msg("Google Sheets handler context cancelled, stopping pull loops")
				return
			}
		}
	}()
}

func (h *Handler) GetCountdown() *types.SimpleCountdown {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	countdownCopy := types.SimpleCountdown{}
	if h.countdown != nil {
		countdownCopy = *h.countdown
	}

	return &countdownCopy
}

func (h *Handler) GetCountdownDuration() *types.SimpleCountdown {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	countdownDurationCopy := types.SimpleCountdown{}
	if h.countdownDuration != nil {
		countdownDurationCopy = *h.countdownDuration
	}

	return &countdownDurationCopy
}

func (h *Handler) GetDJCountdown() *types.DJCountdown {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	djCountdownCopy := types.DJCountdown{}
	if h.djCountdown != nil {
		djCountdownCopy = *h.djCountdown
	}

	return &djCountdownCopy
}

func (h *Handler) GetDJCountdownDuration() *types.DJCountdown {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	djCountdownDurationCopy := types.DJCountdown{}
	if h.djCountdownDuration != nil {
		djCountdownDurationCopy = *h.djCountdownDuration
	}

	return &djCountdownDurationCopy
}

func (h *Handler) GetLowerThird01() *types.LowerThird {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	lowerThirdCopy := types.LowerThird{}
	if h.lowerThird1 != nil {
		lowerThirdCopy = *h.lowerThird1
	}

	return &lowerThirdCopy
}

func (h *Handler) GetLowerThird02() *types.LowerThird {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	lowerThirdCopy := types.LowerThird{}
	if h.lowerThird2 != nil {
		lowerThirdCopy = *h.lowerThird2
	}

	return &lowerThirdCopy
}

func (h *Handler) GetLowerThirdDJ() *types.LowerThird {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	lowerThirdCopy := types.LowerThird{}
	if h.lowerThirdDJ != nil {
		lowerThirdCopy = *h.lowerThirdDJ
	}

	return &lowerThirdCopy
}

func (h *Handler) GetCurrentSchedule() []*types.ScheduleRow {
	scheduleCopy := make([]*types.ScheduleRow, len(h.schedule))
	h.scheduleMtx.RLock()
	copy(scheduleCopy, h.schedule)
	h.scheduleMtx.RUnlock()

	return scheduleCopy
}

func (h *Handler) pullCgSheet() {
	ranges := []string{
		"Main!A3:B3",       // L3 01
		"Main!A7:B7",       // L3 02
		"Main!A11:B11",     // L3 DJ
		"Main!D3:E3",       // General Countdown
		"Main!D7:E7",       // General Countdown from Duration
		"Main!D11:G11",     // DJ Countdown
		"Main!D15:G15",     // DJ Countdown from Duration
		"Schedule!A2:C100", // Schedule
	}

	call := h.sheets.Values.BatchGet(h.cgID).Ranges(ranges...)
	call.DateTimeRenderOption("SERIAL_NUMBER")
	resp, err := call.Do()
	if err != nil {
		h.logger.Error().Err(err).Msg("Unable to retrieve data from CG sheet")
		return
	}

	h.mtx.Lock()
	defer h.mtx.Unlock()

	// Extract L3 01
	if len(resp.ValueRanges) > 0 && len(resp.ValueRanges[0].Values) > 0 {
		values := resp.ValueRanges[0].Values
		if len(values) >= 1 {
			if mainInfo, ok := values[0][0].(string); ok && mainInfo != "" {
				h.lowerThird1.Row1 = mainInfo
			}
			if sideInfo, ok := values[1][0].(string); ok && sideInfo != "" {
				h.lowerThird1.Row2 = sideInfo
			}
		}
	}

	// Extract L3 02
	if len(resp.ValueRanges) > 1 && len(resp.ValueRanges[1].Values) > 0 {
		values := resp.ValueRanges[1].Values
		if len(values) >= 1 {
			if mainInfo, ok := values[0][0].(string); ok && mainInfo != "" {
				h.lowerThird1.Row1 = mainInfo
			}
			if sideInfo, ok := values[1][0].(string); ok && sideInfo != "" {
				h.lowerThird1.Row2 = sideInfo
			}
		}
	}

	// Extract L3 DJ
	if len(resp.ValueRanges) > 2 && len(resp.ValueRanges[2].Values) > 0 {
		values := resp.ValueRanges[2].Values
		if len(values) >= 1 {
			if mainInfo, ok := values[0][0].(string); ok && mainInfo != "" {
				h.lowerThird1.Row1 = mainInfo
			}
			if sideInfo, ok := values[1][0].(string); ok && sideInfo != "" {
				h.lowerThird1.Row2 = sideInfo
			}
		}
	}

	// Extract General Countdown
	if len(resp.ValueRanges) > 3 && len(resp.ValueRanges[3].Values) > 0 {
		countdown, err := extractCountdown(resp.ValueRanges[3].Values, types.CountdownTypeToTime)
		if err != nil {
			h.logger.Error().Err(err).Msg("Unable to extract General Countdown")
		}
		h.countdown = countdown
	}

	// Extract General Countdown Duration
	if len(resp.ValueRanges) > 4 && len(resp.ValueRanges[4].Values) > 0 {
		countdownDuration, err := extractCountdown(resp.ValueRanges[4].Values, types.CountdownTypeDuration)
		if err != nil {
			h.logger.Error().Err(err).Msg("Unable to extract General Countdown Duration")
		}
		h.countdownDuration = countdownDuration
	}

	// Extract DJ Countdown
	if len(resp.ValueRanges) > 5 && len(resp.ValueRanges[5].Values) > 0 {
		djCountdown, err := extractDJCountdown(resp.ValueRanges[5].Values, types.CountdownTypeToTime)
		if err != nil {
			h.logger.Error().Err(err).Msg("Unable to extract DJ Countdown")
		}
		h.djCountdown = djCountdown
	}

	// Extract DJ Countdown Duration
	if len(resp.ValueRanges) > 6 && len(resp.ValueRanges[6].Values) > 0 {
		djCountdownDuration, err := extractDJCountdown(resp.ValueRanges[6].Values, types.CountdownTypeDuration)
		if err != nil {
			h.logger.Error().Err(err).Msg("Unable to extract DJ Countdown Duration")
		}
		h.djCountdownDuration = djCountdownDuration
	}

	// Extract Schedule
	if len(resp.ValueRanges) > 7 && len(resp.ValueRanges[7].Values) > 0 {
		h.schedule = make([]*types.ScheduleRow, 0)
		for i, row := range resp.ValueRanges[7].Values {
			if len(row) < 4 {
				h.logger.Warn().Msgf("Skipping row %d in schedule: expected at least 4 columns, but got %d", i+1, len(row))
				continue
			}

			startTimeSerial, ok1 := row[0].(float64)
			endTimeSerial, ok2 := row[1].(float64)
			title, ok3 := row[2].(string)
			genre, ok4 := row[3].(string)

			if !ok1 || !ok2 || !ok3 || !ok4 {
				h.logger.Warn().Msgf("Skipping row %d in schedule: data has incorrect type", i+1)
				continue
			}

			startTime := parseSerialDateTime(startTimeSerial)
			endTime := parseSerialDateTime(endTimeSerial)

			h.schedule = append(h.schedule, &types.ScheduleRow{
				Title:     title,
				Genre:     genre,
				StartTime: startTime,
				EndTime:   endTime,
			})

		}
	}
}

//
// Helper functions
//

func parseSerialDateTime(serialDate float64) time.Time {
	// Google Sheets' epoch starts on 1899-12-30.
	excelEpoch := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)

	// Separate the integer part (days) from the fractional part (time).
	days := math.Floor(serialDate)
	fraction := serialDate - days

	// The 1900 leap year bug in Excel/Sheets: They incorrectly believe
	// 1900 was a leap year. Serial dates >= 61 (representing March 1, 1900)
	// are off by one. We must subtract a day to compensate.
	if days >= 61 {
		days--
	}

	datePart := excelEpoch.AddDate(0, 0, int(days))
	dayDuration := time.Duration(24 * time.Hour)
	timeFraction := time.Duration(fraction * float64(dayDuration))

	return datePart.Add(timeFraction)
}

func extractCountdown(values [][]interface{}, countdownType types.CountdownType) (*types.SimpleCountdown, error) {
	if len(values) >= 1 {
		var info, countdownTime string = "Countdown", "00:00:00"

		if len(values) > 0 && len(values[0]) > 0 {
			if mainInfo, ok := values[0][0].(string); ok && mainInfo != "" {
				info = mainInfo
			}
		}
		if len(values) > 1 && len(values[1]) > 0 {
			if countdownInfo, ok := values[1][0].(string); ok && countdownInfo != "" {
				countdownTime = countdownInfo
			}
		}

		return &types.SimpleCountdown{
			CountdownType: countdownType,
			Text:          info,
			CountdownTime: countdownTime,
		}, nil
	}

	return nil, fmt.Errorf("not enough data to extract Countdown")
}

func extractDJCountdown(values [][]interface{}, countdownType types.CountdownType) (*types.DJCountdown, error) {
	if len(values) >= 1 {
		var info, countdownTime, dj1, dj2, genre string = "Countdown", "00:00:00", "", "", ""

		if len(values) > 0 && len(values[0]) > 0 {
			if mainInfo, ok := values[0][0].(string); ok && mainInfo != "" {
				info = mainInfo
			}
		}
		if len(values) > 1 && len(values[1]) > 0 {
			if countdownInfo, ok := values[1][0].(string); ok && countdownInfo != "" {
				countdownTime = countdownInfo
			}
		}
		if len(values) > 2 && len(values[2]) > 0 {
			if dj1Info, ok := values[2][0].(string); ok && dj1Info != "" {
				dj1 = dj1Info
			}
		}
		if len(values) > 3 && len(values[3]) > 0 {
			if dj2Info, ok := values[3][0].(string); ok && dj2Info != "" {
				dj2 = dj2Info
			}
		}
		if len(values) > 4 && len(values[4]) > 0 {
			if genreInfo, ok := values[4][0].(string); ok && genreInfo != "" {
				genre = genreInfo
			}
		}

		return &types.DJCountdown{
			SimpleCountdown: &types.SimpleCountdown{
				CountdownType: countdownType,
				Text:          info,
				CountdownTime: countdownTime,
			},
			FirstDJ:  dj1,
			SecondDJ: dj2,
			Genre:    genre,
		}, nil
	}

	return nil, fmt.Errorf("not enough data to extract DJ Countdown")
}

func (h *Handler) Close() {
	h.logger.Info().Msg("Stopping Google Sheets handler")

	h.cancel()
	h.wg.Wait()
}
