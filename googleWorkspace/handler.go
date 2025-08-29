package googleworkspace

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/OverlayFox/CaspawCG/googleWorkspace/types"
	gTypes "github.com/OverlayFox/CaspawCG/types"
	"github.com/rs/zerolog"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Handler struct {
	logger zerolog.Logger

	sheets *sheets.SpreadsheetsService
	cgID   string

	countdown           *types.SheetsCountdown
	countdownDuration   *types.SheetsCountdown
	djCountdown         *types.SheetsDJCountdown
	djCountdownDuration *types.SheetsDJCountdown
	lowerThird1         *types.SheetsLowerThird
	lowerThird2         *types.SheetsLowerThird
	lowerThirdDJ        *types.SheetsLowerThird
	schedule            []*types.SheetsScheduleRow

	mtx         sync.RWMutex
	scheduleMtx sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewHandler(logger zerolog.Logger) (gTypes.SheetsHandler, error) {
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

		countdownDuration: &types.SheetsCountdown{
			CountdownType: types.CountdownTypeDuration,
			Text:          "Countdown",
			CountdownTime: "00:00:00",
		},
		countdown: &types.SheetsCountdown{
			CountdownType: types.CountdownTypeToTime,
			Text:          "Countdown",
			CountdownTime: "00:00:00",
		},
		djCountdownDuration: &types.SheetsDJCountdown{
			CountdownType: types.CountdownTypeDuration,
			Name:          "DJ Countdown",
			Genre:         "<DJ Genre>",
			CountdownTime: "00:00:00",
		},
		djCountdown: &types.SheetsDJCountdown{
			CountdownType: types.CountdownTypeToTime,
			Name:          "DJ Countdown",
			Genre:         "<DJ Genre>",
			CountdownTime: "00:00:00",
		},
		lowerThird1: &types.SheetsLowerThird{
			Row1: "",
			Row2: "",
		},
		lowerThird2: &types.SheetsLowerThird{
			Row1: "",
			Row2: "",
		},
		schedule: make([]*types.SheetsScheduleRow, 0),
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

func (h *Handler) GetData(layer int, templatePath gTypes.TemplatePaths) ([]byte, error) {
	switch templatePath {
	case gTypes.TemplatePathCountdown, gTypes.TemplatePathCountdownToTime:
		var sheetsCountdown *types.SheetsCountdown
		switch layer {
		case 20:
			sheetsCountdown = h.getCountdown()
		case 21:
			sheetsCountdown = h.getCountdownDuration()
		default:
			return nil, fmt.Errorf("unknown layer '%d' for countdown template", layer)
		}
		countdown := gTypes.Countdown{
			Title:        sheetsCountdown.Text,
			TimerMinutes: "00:00",
			TimerHours:   sheetsCountdown.CountdownTime,
		}
		return countdown.MarshalJSON()

	case gTypes.TemplatePathDJCountdown, gTypes.TemplatePathDJCountdownToTime:
		var sheetsCountdown *types.SheetsDJCountdown
		switch layer {
		case 30:
			sheetsCountdown = h.getDJCountdown()
		case 31:
			sheetsCountdown = h.getDJCountdownDuration()
		default:
			return nil, fmt.Errorf("unknown layer '%d' for dj-countdown template", layer)
		}
		logo, err := getDJLogo(sheetsCountdown.Name)
		if err != nil {
			return nil, err
		}
		countdown := gTypes.DJCountdown{
			Name:         sheetsCountdown.Name,
			Genre:        sheetsCountdown.Genre,
			Logo:         logo,
			TimerMinutes: "00:00",
			TimerHours:   sheetsCountdown.CountdownTime,
		}
		return countdown.MarshalJSON()

	case gTypes.TemplatePathSchedule:
		return h.GetCurrentSchedule(0)
	}

	return nil, fmt.Errorf("unknown template path: %s", templatePath)
}

func (h *Handler) getCountdown() *types.SheetsCountdown {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	countdownCopy := types.SheetsCountdown{}
	if h.countdown != nil {
		countdownCopy = *h.countdown
	}

	return &countdownCopy
}

func (h *Handler) getCountdownDuration() *types.SheetsCountdown {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	countdownDurationCopy := types.SheetsCountdown{}
	if h.countdownDuration != nil {
		countdownDurationCopy = *h.countdownDuration
	}

	return &countdownDurationCopy
}

func (h *Handler) getDJCountdown() *types.SheetsDJCountdown {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	djCountdownCopy := types.SheetsDJCountdown{}
	if h.djCountdown != nil {
		djCountdownCopy = *h.djCountdown
	}

	return &djCountdownCopy
}

func (h *Handler) getDJCountdownDuration() *types.SheetsDJCountdown {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	djCountdownDurationCopy := types.SheetsDJCountdown{}
	if h.djCountdownDuration != nil {
		djCountdownDurationCopy = *h.djCountdownDuration
	}

	return &djCountdownDurationCopy
}

func (h *Handler) getLowerThird01() *types.SheetsLowerThird {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	lowerThirdCopy := types.SheetsLowerThird{}
	if h.lowerThird1 != nil {
		lowerThirdCopy = *h.lowerThird1
	}

	return &lowerThirdCopy
}

func (h *Handler) getLowerThird02() *types.SheetsLowerThird {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	lowerThirdCopy := types.SheetsLowerThird{}
	if h.lowerThird2 != nil {
		lowerThirdCopy = *h.lowerThird2
	}

	return &lowerThirdCopy
}

func (h *Handler) getLowerThirdDJ() *types.SheetsLowerThird {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	lowerThirdCopy := types.SheetsLowerThird{}
	if h.lowerThirdDJ != nil {
		lowerThirdCopy = *h.lowerThirdDJ
	}

	return &lowerThirdCopy
}

func (h *Handler) GetCurrentSchedule(startIndex int) ([]byte, error) {
	scheduleCopy := make([]*types.SheetsScheduleRow, len(h.schedule))
	h.scheduleMtx.RLock()
	copy(scheduleCopy, h.schedule)
	h.scheduleMtx.RUnlock()

	result := make([]*types.SheetsScheduleRow, 0, 3)
	n := len(scheduleCopy)
	if n == 0 {
		return nil, nil
	}
	for i := 0; i < 3; i++ {
		idx := (startIndex + i) % n
		row := scheduleCopy[idx]
		if row == nil {
			continue
		}
		result = append(result, &types.SheetsScheduleRow{
			Name:      row.Name,
			Genre:     row.Genre,
			StartTime: row.StartTime,
		})
	}

	logos := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		logo, err := getDJLogo(result[i].Name)
		if err != nil {
			return nil, err
		}
		logos = append(logos, logo)
	}

	schedule := gTypes.Schedule{
		Name01: result[0].Name,
		Name02: result[1].Name,
		Name03: result[2].Name,

		Genre01: result[0].Genre,
		Genre02: result[1].Genre,
		Genre03: result[2].Genre,

		Time01: result[0].StartTime.Format("15:04"),
		Time02: result[1].StartTime.Format("15:04"),
		Time03: result[2].StartTime.Format("15:04"),

		Logo01: logos[0],
		Logo02: logos[1],
		Logo03: logos[2],
	}

	return schedule.MarshalJSON()
}

func (h *Handler) pullCgSheet() {
	ranges := []string{
		"Main!A3:B3",       // L3 01
		"Main!A7:B7",       // L3 02
		"Main!A11:B11",     // L3 DJ
		"Main!D3:E3",       // General Countdown
		"Main!D7:E7",       // General Countdown from Duration
		"Main!D11:F11",     // DJ Countdown
		"Main!D15:F15",     // DJ Countdown from Duration
		"Schedule!A2:C100", // Schedule
	}

	call := h.sheets.Values.BatchGet(h.cgID).Ranges(ranges...).ValueRenderOption("UNFORMATTED_VALUE").DateTimeRenderOption("SERIAL_NUMBER")
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
		if len(values) >= 2 {
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
		if len(values) >= 2 {
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
		if len(values) >= 2 {
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
		h.schedule = make([]*types.SheetsScheduleRow, 0)
		for i, row := range resp.ValueRanges[7].Values {
			if len(row) < 3 {
				h.logger.Warn().Msgf("Skipping row %d in schedule: expected at least 4 columns, but got %d", i+1, len(row))
				continue
			}

			startTimeSerial, ok1 := row[0].(float64)
			name, ok2 := row[1].(string)
			genre, ok3 := row[2].(string)

			if !ok1 || !ok2 || !ok3 {
				h.logger.Warn().Msgf("Skipping row %d in schedule: data has incorrect type", i+1)
				continue
			}

			startTime, err := parseSerialDateTime(startTimeSerial)
			if err != nil {
				h.logger.Warn().Err(err).Msgf("Skipping row %d in schedule: invalid start time", i+1)
				continue
			}

			// h.logger.Debug().Str("name", name).Str("genre", genre).Time("start_time", startTime).Msg("Extracted schedule row")

			h.schedule = append(h.schedule, &types.SheetsScheduleRow{
				Name:      name,
				Genre:     genre,
				StartTime: startTime,
			})

		}
	}
}

//
// Helper functions
//

func parseSerialDateTime(serialDate float64) (time.Time, error) {
	// Google Sheets' epoch starts on 1899-12-30.
	berlinLoc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		return time.Time{}, err
	}
	excelEpoch := time.Date(1899, 12, 30, 0, 0, 0, 0, berlinLoc)

	// Separate the integer part (days) from the fractional part (time).
	days := math.Floor(serialDate)
	fraction := serialDate - days

	datePart := excelEpoch.AddDate(0, 0, int(days))
	dayDuration := time.Duration(24 * time.Hour)
	timeFraction := time.Duration(fraction * float64(dayDuration))

	return datePart.Add(timeFraction), nil
}

func extractCountdown(values [][]any, countdownType types.CountdownType) (*types.SheetsCountdown, error) {
	if len(values) >= 1 {
		var info, countdownTime string = "Countdown", "00:00:00"

		if len(values[0]) > 0 {
			if time, ok := values[0][0].(string); ok && time != "" {
				countdownTime = time
			}
		}
		if len(values[0]) > 1 {
			if mainInfo, ok := values[0][1].(string); ok && mainInfo != "" {
				info = mainInfo
			}

		}

		return &types.SheetsCountdown{
			CountdownType: countdownType,
			Text:          info,
			CountdownTime: countdownTime,
		}, nil
	}

	return nil, fmt.Errorf("not enough data to extract Countdown")
}

func extractDJCountdown(values [][]any, countdownType types.CountdownType) (*types.SheetsDJCountdown, error) {
	if len(values) >= 1 {
		var countdownTime, name, genre string = "00:00:00", "", ""

		if len(values[0]) > 0 {
			if countdownInfo, ok := values[0][0].(string); ok && countdownInfo != "" {
				countdownTime = countdownInfo
			}
		}
		if len(values[0]) > 1 {
			if nameInfo, ok := values[0][1].(string); ok && nameInfo != "" {
				name = nameInfo
			}
		}
		if len(values[0]) > 2 {
			if genreInfo, ok := values[0][2].(string); ok && genreInfo != "" {
				genre = genreInfo
			}
		}

		return &types.SheetsDJCountdown{
			CountdownType: countdownType,
			Name:          name,
			Genre:         genre,
			CountdownTime: countdownTime,
		}, nil
	}

	return nil, fmt.Errorf("not enough data to extract DJ Countdown")
}

func (h *Handler) Close() {
	h.logger.Info().Msg("Stopping Google Sheets handler")

	h.cancel()
	h.wg.Wait()
}
