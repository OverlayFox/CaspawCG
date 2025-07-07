package googleworkspace

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/OverlayFox/CaspawCG/types"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Handler struct {
	sheets *sheets.SpreadsheetsService
	id     string

	countdown               *types.Countdown
	countdownToTime         *types.Countdown
	lowerThird              *types.LowerThird
	detailedDanceCompSingle *types.DetailedDanceComp
	detailedDanceComp       []*types.DetailedDanceComp

	mtx sync.RWMutex

	ctx context.Context
	wg  sync.WaitGroup
}

func NewHandler() (types.SheetsData, error) {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("service-account.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to create Sheets service: %w", err)
	}

	return &Handler{
		sheets: srv.Spreadsheets,
		id:     "1ag__aae56azH_D4yqQpXXf4GXpsMespSlju_ed20vVo",
		ctx:    ctx,

		countdown: &types.Countdown{
			Title:         "Countdown",
			CountdownTime: "00:00:00",
		},
		countdownToTime: &types.Countdown{
			Title:         "Countdown",
			CountdownTime: "00:00:00",
		},
		lowerThird: &types.LowerThird{
			Row1: "",
			Row2: "",
		},
		detailedDanceComp: make([]*types.DetailedDanceComp, 0),
	}, nil
}

func (h *Handler) Start() {
	h.pull()
	return

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				h.pull()
			case <-h.ctx.Done():
				log.Println("Stopping Google Sheets handler")
				return
			}
		}
	}()
}

func (h *Handler) GetCountdown() *types.Countdown {
	h.mtx.RLock()
	defer h.mtx.RUnlock()
	return h.countdown
}

func (h *Handler) GetCountdownToTime() *types.Countdown {
	h.mtx.RLock()
	defer h.mtx.RUnlock()
	return h.countdownToTime
}

func (h *Handler) GetLowerThird() *types.LowerThird {
	h.mtx.RLock()
	defer h.mtx.RUnlock()
	return h.lowerThird
}

func (h *Handler) GetDetailedDanceCompSingle() *types.DetailedDanceComp {
	h.mtx.RLock()
	defer h.mtx.RUnlock()
	return h.detailedDanceCompSingle
}

func (h *Handler) GetDetailedDanceComp() []*types.DetailedDanceComp {
	h.mtx.RLock()
	defer h.mtx.RUnlock()
	return h.detailedDanceComp
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

	log.Printf("Expected string value for %s, got: %v", fieldName, valueRange.Values[0][0])
	return ""
}

func (h *Handler) pull() {
	resp, err := h.sheets.Values.BatchGet(h.id).Ranges("Main!B2:B3", "Main!B6:B7", "Main!F2:F3", "DanceComp!A2:F89", "Main!B10").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
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

	// Extract countdownToTime value (B5)
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
				h.lowerThird.Row1 = row1
			}
		}
		if len(values) >= 2 {
			if row2, ok := values[1][0].(string); ok {
				h.lowerThird.Row2 = row2
			}
		}
	}

	// Extract dance comp values (DanceComp!A2:DanceComp!F89)
	if len(resp.ValueRanges) > 3 && len(resp.ValueRanges[3].Values) > 0 {
		values := resp.ValueRanges[3].Values

		// Helper function to safely extract string value from cell
		getValue := func(row, col int) string {
			if row < len(values) && col < len(values[row]) && col < len(values[row]) {
				if val, ok := values[row][col].(string); ok && val != "" {
					return val
				}
			}
			return "0"
		}

		// Process data in chunks of 8 rows per participant
		for i := 0; i < len(values)-7; i += 8 {
			name := getValue(i, 0)
			if name == "0" {
				name = "Unknown"
			}

			h.detailedDanceComp = append(h.detailedDanceComp, &types.DetailedDanceComp{
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
	}

	// Sort h.detailedDanceComp by TotalScore descending
	sort.Slice(h.detailedDanceComp, func(i, j int) bool {
		var scoreI, scoreJ float64
		fmt.Sscanf(h.detailedDanceComp[i].TotalScore, "%f", &scoreI)
		fmt.Sscanf(h.detailedDanceComp[j].TotalScore, "%f", &scoreJ)
		return scoreI > scoreJ
	})

	// Extract selected dance comp participant (Main!B8)
	if danceCompName := h.extractStringValue(resp, 4, "dance competitor"); danceCompName != "" {
		for _, comp := range h.detailedDanceComp {
			if comp.Name == danceCompName {
				h.detailedDanceCompSingle = comp
				break
			}
		}
	}
}
