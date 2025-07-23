package googleworkspace

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/OverlayFox/CaspawCG/types"
	"github.com/rs/zerolog"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Handler struct {
	logger zerolog.Logger

	sheets     *sheets.SpreadsheetsService
	cgID       string
	scheduleID string

	freeStandings           []*types.FreeStandings
	countdown               *types.Countdown
	countdownToTime         *types.Countdown
	lowerThird1             *types.LowerThird
	lowerThird2             *types.LowerThird
	lowerThird3             *types.LowerThird
	detailedDanceCompSingle *types.DetailedDanceComp
	detailedDanceComp       []*types.DetailedDanceComp
	schedule                []*types.ScheduleRow
	attribution             map[string]string // Key is the contestants name, value is the person who needs to be credited

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

		sheets:     srv.Spreadsheets,
		cgID:       "1ag__aae56azH_D4yqQpXXf4GXpsMespSlju_ed20vVo",
		scheduleID: "1NZbShItG_jQRS48_PXtE0dRnQp0NqOrn__I6fxRpiws",
		ctx:        ctx,
		cancel:     cancel,

		countdown: &types.Countdown{
			Title:         "Countdown",
			CountdownTime: "00:00:00",
		},
		countdownToTime: &types.Countdown{
			Title:         "Countdown",
			CountdownTime: "00:00:00",
		},
		lowerThird1: &types.LowerThird{
			Row1: "",
			Row2: "",
		},
		lowerThird2: &types.LowerThird{
			Row1: "",
			Row2: "",
		},
		lowerThird3: &types.LowerThird{
			Row1: "",
			Row2: "",
		},
		detailedDanceComp: make([]*types.DetailedDanceComp, 0),
		schedule:          make([]*types.ScheduleRow, 0),
		attribution:       make(map[string]string),
	}, nil
}

func (h *Handler) Start() {
	h.logger.Info().Msg("Starting Google Sheets handler")

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()

		h.logger.Debug().Msg("Pulling initial data from Google Sheets")
		h.pullCgSheet()
		h.pullSchedule()

		ticker5sec := time.NewTicker(5 * time.Second)
		defer ticker5sec.Stop()

		ticker5min := time.NewTicker(5 * time.Minute)
		defer ticker5min.Stop()

		for {
			select {
			case <-ticker5sec.C:
				h.pullCgSheet()
			case <-ticker5min.C:
				h.logger.Trace().Msg("Pulling schedule data")
				h.pullSchedule()
			case <-h.ctx.Done():
				h.logger.Debug().Msg("Google Sheets handler context cancelled, stopping pull loops")
				return
			}
		}
	}()
}

func (h *Handler) GetCountdown() *types.Countdown {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	countdownCopy := types.Countdown{}
	if h.countdown != nil {
		countdownCopy = *h.countdown
	}

	return &countdownCopy
}

func (h *Handler) GetCountdownToTime() *types.Countdown {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	countdownToTimeCopy := types.Countdown{}
	if h.countdownToTime != nil {
		countdownToTimeCopy = *h.countdownToTime
	}

	return &countdownToTimeCopy
}

func (h *Handler) GetLowerThirdSingle() *types.LowerThird {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	lowerThirdCopy := types.LowerThird{}
	if h.lowerThird1 != nil {
		lowerThirdCopy = *h.lowerThird1
	}

	return &lowerThirdCopy
}

func (h *Handler) GetLowerThirdDuo() (*types.LowerThird, *types.LowerThird) {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	lowerThirdCopy2 := types.LowerThird{}
	if h.lowerThird2 != nil {
		lowerThirdCopy2 = *h.lowerThird2
	}

	lowerThirdCopy3 := types.LowerThird{}
	if h.lowerThird3 != nil {
		lowerThirdCopy3 = *h.lowerThird3
	}

	return &lowerThirdCopy2, &lowerThirdCopy3
}

func (h *Handler) GetDetailedDanceCompSingle() *types.DetailedDanceComp {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	detailedDanceCompCopy := types.DetailedDanceComp{}
	if h.detailedDanceCompSingle != nil {
		detailedDanceCompCopy = *h.detailedDanceCompSingle
	}

	return &detailedDanceCompCopy
}

func (h *Handler) GetDetailedDanceComp() []*types.DetailedDanceComp {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	detailedDanceCompCopy := make([]*types.DetailedDanceComp, len(h.detailedDanceComp))
	copy(detailedDanceCompCopy, h.detailedDanceComp)

	return detailedDanceCompCopy
}

func (h *Handler) GetCurrentSchedule() []*types.ScheduleRow {
	scheduleCopy := make([]*types.ScheduleRow, len(h.schedule))
	h.scheduleMtx.RLock()
	copy(scheduleCopy, h.schedule)
	h.scheduleMtx.RUnlock()

	currentSchedule := make([]*types.ScheduleRow, 0)
	now := time.Now().Add(10 * time.Minute) // Only show events that aren't about to be over
	for _, row := range scheduleCopy {
		if now.After(row.EndTime) {
			continue
		}
		currentSchedule = append(currentSchedule, row)
	}
	return currentSchedule
}

func (h *Handler) GetAttribution(contestantsName string) (string, error) {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	if attribution, ok := h.attribution[contestantsName]; ok {
		return attribution, nil
	}
	return "", fmt.Errorf("no attribution found for contestant: %s", contestantsName)
}

func (h *Handler) GetFreeStandings() []*types.FreeStandings {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	freeStandingsCopy := make([]*types.FreeStandings, len(h.freeStandings))
	copy(freeStandingsCopy, h.freeStandings)

	return freeStandingsCopy
}

func (h *Handler) extractStringValue(resp *sheets.BatchGetValuesResponse, rangeIndex int, fieldName string) string {
	if len(resp.ValueRanges) <= rangeIndex {
		return ""
	}

	valueRange := resp.ValueRanges[rangeIndex]
	if len(valueRange.Values) == 0 || len(valueRange.Values[0]) == 0 {
		return ""
	}

	if val, ok := valueRange.Values[0][0].(string); ok {
		return val
	}

	h.logger.Warn().Msgf("Expected string value for '%s', got: '%v'", fieldName, valueRange.Values[0][0])
	return ""
}

func (h *Handler) pullSchedule() {
	resp, err := h.sheets.Values.BatchGet(h.scheduleID).Ranges("B2:B77", "E2:E77", "F2:H77").Do()
	if err != nil {
		h.logger.Error().Err(err).Msg("Unable to retrieve data from schedule sheet")
		return
	}

	if len(resp.ValueRanges) < 3 {
		h.logger.Warn().Msg("Not enough data ranges returned from the schedule sheet")
		return
	}

	titles := resp.ValueRanges[0].Values
	rooms := resp.ValueRanges[1].Values
	times := resp.ValueRanges[2].Values
	scheduleList := make([]*types.ScheduleRow, 0, len(titles))

	blacklist := []string{"Pickup Payment", "Crew", "[STREAMING]"}

	for i := 0; i < len(titles); i++ {
		var title, room, weekDay, startTime, duration string

		if len(titles) > i && len(titles[i]) > 0 {
			if val, ok := titles[i][0].(string); ok {
				title = strings.TrimSpace(val)
				lowerTitle := strings.ToLower(title)
				blacklisted := false
				for _, word := range blacklist {
					if strings.Contains(lowerTitle, strings.ToLower(word)) {
						blacklisted = true
						break
					}
				}
				if blacklisted {
					continue // Skip this row if blacklisted
				}
			}
		}

		if len(rooms) > i && len(rooms[i]) > 0 {
			if val, ok := rooms[i][0].(string); ok {
				room = strings.TrimSpace(val)
			}
		}

		if len(times) > i && len(times[i]) > 2 {
			if val, ok := times[i][0].(string); ok {
				weekDay = strings.TrimSpace(val)
			}
			if val, ok := times[i][1].(string); ok {
				startTime = strings.TrimSpace(val)
			}
			if val, ok := times[i][2].(string); ok {
				duration = strings.TrimSpace(val)
			}
		}

		schedule, err := types.NewScheduleRow(title, room, weekDay, startTime, duration)
		if err != nil {
			h.logger.Error().Err(err).Msgf("Error creating schedule row for index '%d'", i)
			continue
		}
		scheduleList = append(scheduleList, schedule)
	}

	// Sort scheduleList by StartTime ascending (earliest first)
	sort.Slice(scheduleList, func(i, j int) bool {
		if scheduleList[i].StartTime.Before(scheduleList[j].StartTime) {
			return true
		}
		if scheduleList[j].StartTime.Before(scheduleList[i].StartTime) {
			return false
		}
		return scheduleList[i].EndTime.Before(scheduleList[j].EndTime) // If start times are equal, compare by end time
	})

	h.scheduleMtx.Lock()
	h.schedule = scheduleList
	h.scheduleMtx.Unlock()
}

func (h *Handler) pullCgSheet() {
	resp, err := h.sheets.Values.BatchGet(h.cgID).Ranges("Main!B2:B3", "Main!B6:B7", "Main!F2:F3", "Main!F6:F7", "Main!H6:H7", "DanceComp!A2:F89", "Main!B10", "Lists!A2:B12", "Standings!A2:B13").Do()
	if err != nil {
		h.logger.Error().Err(err).Msg("Unable to retrieve data from CG sheet")
		return
	}

	h.mtx.Lock()
	defer h.mtx.Unlock()

	// Extract countdown value (B2)
	if len(resp.ValueRanges) > 0 && len(resp.ValueRanges[0].Values) > 0 {
		values := resp.ValueRanges[0].Values
		if len(values) >= 1 {
			if title, ok := values[0][0].(string); ok && title != "" {
				h.countdown.Title = title
			}
			if timer, ok := values[1][0].(string); ok && timer != "" {
				h.countdown.CountdownTime = timer
			}
		}
	}

	// Extract countdownToTime value (B6)
	if len(resp.ValueRanges) > 1 && len(resp.ValueRanges[1].Values) > 0 {
		values := resp.ValueRanges[1].Values
		if len(values) >= 1 {
			if title, ok := values[0][0].(string); ok && title != "" {
				h.countdownToTime.Title = title
			}
			if timer, ok := values[1][0].(string); ok && timer != "" {
				h.countdownToTime.CountdownTime = timer
			}
		}
	}

	// Extract lower third values (F2:F3)
	if len(resp.ValueRanges) > 2 && len(resp.ValueRanges[2].Values) > 0 {
		values := resp.ValueRanges[2].Values
		if len(values) >= 1 {
			if row1, ok := values[0][0].(string); ok {
				h.lowerThird1.Row1 = row1
			}
		}
		if len(values) >= 2 {
			if row2, ok := values[1][0].(string); ok {
				h.lowerThird1.Row2 = row2
			}
		}
	}

	// Extract lower third values (F6:F7)
	if len(resp.ValueRanges) > 3 && len(resp.ValueRanges[3].Values) > 0 {
		values := resp.ValueRanges[3].Values
		if len(values) >= 1 {
			if row1, ok := values[0][0].(string); ok {
				h.lowerThird2.Row1 = row1
			}
		}
		if len(values) >= 2 {
			if row2, ok := values[1][0].(string); ok {
				h.lowerThird2.Row2 = row2
			}
		}
	}

	// Extract lower third values (H6:H7)
	if len(resp.ValueRanges) > 4 && len(resp.ValueRanges[4].Values) > 0 {
		values := resp.ValueRanges[4].Values
		if len(values) >= 1 {
			if row1, ok := values[0][0].(string); ok {
				h.lowerThird3.Row1 = row1
			}
		}
		if len(values) >= 2 {
			if row2, ok := values[1][0].(string); ok {
				h.lowerThird3.Row2 = row2
			}
		}
	}

	// Extract dance comp values (DanceComp!A2:DanceComp!F89)
	if len(resp.ValueRanges) > 5 && len(resp.ValueRanges[5].Values) > 0 {
		values := resp.ValueRanges[5].Values

		// Helper function to safely extract string value from cell
		getValue := func(row, col int) string {
			if row < len(values) && col < len(values[row]) && col < len(values[row]) {
				if val, ok := values[row][col].(string); ok && val != "" {
					return val
				}
			}
			return "0"
		}

		detailedDanceComp := make([]*types.DetailedDanceComp, 0)

		// Process data in chunks of 8 rows per participant
		for i := 0; i < len(values)-7; i += 8 {
			name := getValue(i, 0)
			if name == "0" {
				name = "Unknown"
			}

			detailedDanceComp = append(detailedDanceComp, &types.DetailedDanceComp{
				Name:            name,
				TotalScore:      getValue(i, 5),
				Appearance:      getValue(i+1, 5),
				Professionalism: getValue(i+2, 5),
				Consistency:     getValue(i+3, 5),
				Complexity:      getValue(i+4, 5),
				Decibels:        getValue(i+5, 5),
				Originality:     getValue(i+6, 5),
				Quantum:         getValue(i+7, 5),
			})
		}

		// Sort h.detailedDanceComp by TotalScore descending
		if len(detailedDanceComp) > 0 {
			sort.Slice(detailedDanceComp, func(i, j int) bool {
				if detailedDanceComp[i] == nil {
					return false // nil goes to the back
				}
				if detailedDanceComp[j] == nil {
					return true // non-nil comes before nil
				}

				// Handle nil or empty TotalScore
				if detailedDanceComp[i].TotalScore == "" || detailedDanceComp[i].TotalScore == "0" {
					if detailedDanceComp[j].TotalScore == "" || detailedDanceComp[j].TotalScore == "0" {
						return false // both are empty/zero, keep order
					}
					return false // i goes after j
				}
				if detailedDanceComp[j].TotalScore == "" || detailedDanceComp[j].TotalScore == "0" {
					return true // i goes before j
				}

				var scoreI, scoreJ float64
				fmt.Sscanf(detailedDanceComp[i].TotalScore, "%f", &scoreI)
				fmt.Sscanf(detailedDanceComp[j].TotalScore, "%f", &scoreJ)
				return scoreI > scoreJ
			})
		}
		h.detailedDanceComp = detailedDanceComp
	}

	// Extract selected dance comp participant (Main!B8)
	if danceCompName := h.extractStringValue(resp, 6, "dance competitor"); danceCompName != "" {
		for _, comp := range h.detailedDanceComp {
			if comp.Name == danceCompName {
				h.detailedDanceCompSingle = comp
				break
			}
		}
	}

	// Extract attribution data (Lists!A2:B12)
	if len(resp.ValueRanges) > 7 && len(resp.ValueRanges[7].Values) > 0 {
		values := resp.ValueRanges[7].Values
		h.attribution = make(map[string]string)

		for _, row := range values {
			if len(row) < 2 {
				continue // Skip rows that don't have at least 2 columns
			}
			name, okName := row[0].(string)
			attribution, okAttribution := row[1].(string)
			if okName && okAttribution && name != "" && attribution != "" {
				h.attribution[name] = attribution
			}
		}
	}

	// Extract free standings data (Standings!A2:B13)
	if len(resp.ValueRanges) > 8 && len(resp.ValueRanges[8].Values) > 0 {
		values := resp.ValueRanges[8].Values
		h.freeStandings = make([]*types.FreeStandings, 0)

		for _, row := range values {
			if len(row) < 2 {
				continue // Skip rows that don't have at least 2 columns
			}
			points := "0"
			pointsStr, okPoints := row[0].(string)
			if okPoints {
				points = strings.TrimSpace(pointsStr)
			} else {
				h.logger.Error().Err(err).Msgf("Invalid points value: %s", pointsStr)
			}

			nameStr, okName := row[1].(string)
			if okName && nameStr != "" {
				h.freeStandings = append(h.freeStandings, &types.FreeStandings{
					ContestantName: strings.TrimSpace(nameStr),
					Points:         points,
				})
			}
		}
	}
}

func (h *Handler) Close() {
	h.logger.Info().Msg("Stopping Google Sheets handler")

	h.cancel()
	h.wg.Wait()
}
