package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type TemplatePaths string

const (
	TemplatePathCountdown         TemplatePaths = "COUNTDOWN_DURATION"
	TemplatePathCountdownToTime   TemplatePaths = "COUNTDOWN_TO_TIME"
	TemplatePathDJCountdown       TemplatePaths = "DJ-COUNTDOWN_DURATION"
	TemplatePathDJCountdownToTime TemplatePaths = "DJ-COUNTDOWN_TO_TIME"
	TemplatePathTitle             TemplatePaths = "TITLE"
	TemplatePathSchedule          TemplatePaths = "SCHEDULE"
	TemplatePathLowerThird        TemplatePaths = "L3"
)

var commandTemplatePathMap = map[string]TemplatePaths{
	"COUNTDOWN_DURATION":    TemplatePathCountdown,
	"COUNTDOWN_TO_TIME":     TemplatePathCountdownToTime,
	"DJ-COUNTDOWN_DURATION": TemplatePathDJCountdown,
	"DJ-COUNTDOWN_TO_TIME":  TemplatePathDJCountdownToTime,
	"TITLE":                 TemplatePathTitle,
	"SCHEDULE":              TemplatePathSchedule,
	"L3":                    TemplatePathLowerThird,
}

func TemplatePathFromString(s string) (TemplatePaths, error) {
	if tp, ok := commandTemplatePathMap[strings.ToUpper(s)]; ok {
		return tp, nil
	}
	return "", fmt.Errorf("unknown template path: %s", s)
}

//
// Template Paths Layer Map
//

var commandTemplateLayerMap = map[int]TemplatePaths{
	20: TemplatePathCountdownToTime,
	21: TemplatePathCountdown,
	30: TemplatePathDJCountdownToTime,
	31: TemplatePathDJCountdown,
	40: TemplatePathTitle,
	41: TemplatePathSchedule,
}

func TemplateFromLayer(layer int) (TemplatePaths, error) {
	if tp, ok := commandTemplateLayerMap[layer]; ok {
		return tp, nil
	}
	return "", fmt.Errorf("unknown template for cg_layer: %d", layer)
}

//
// Templates
//

type Countdown struct {
	Title        string `json:"_title"`
	TimerMinutes string `json:"_timerMinutes"`
	TimerHours   string `json:"_timerHours"`
}

func (c *Countdown) MarshalJSON() ([]byte, error) {
	type CountdownAlias Countdown
	return marshalJSON((*CountdownAlias)(c))
}

type DJCountdown struct {
	Name         string `json:"_name"`
	Genre        string `json:"_genre"`
	Logo         string `json:"_logo"`
	TimerMinutes string `json:"_timerMinutes"`
	TimerHours   string `json:"_timerHours"`
}

func (c *DJCountdown) MarshalJSON() ([]byte, error) {
	type DJCountdownAlias DJCountdown
	return marshalJSON((*DJCountdownAlias)(c))
}

type Title struct {
	Title string `json:"_title"`
}

func (c *Title) MarshalJSON() ([]byte, error) {
	type TitleAlias Title
	return marshalJSON((*TitleAlias)(c))
}

type Schedule struct {
	Name01  string `json:"_name01"`
	Genre01 string `json:"_genre01"`
	Time01  string `json:"_time01"`
	Logo01  string `json:"_logo01"`

	Name02  string `json:"_name02"`
	Genre02 string `json:"_genre02"`
	Time02  string `json:"_time02"`
	Logo02  string `json:"_logo02"`

	Name03  string `json:"_name03"`
	Genre03 string `json:"_genre03"`
	Time03  string `json:"_time03"`
	Logo03  string `json:"_logo03"`
}

func (c *Schedule) MarshalJSON() ([]byte, error) {
	type ScheduleAlias Schedule
	return marshalJSON((*ScheduleAlias)(c))
}

type LowerThird struct {
	Name string `json:"_name"`
	Info string `json:"_info"`
}

func (c *LowerThird) MarshalJSON() ([]byte, error) {
	type LowerThirdAlias LowerThird
	return marshalJSON((*LowerThirdAlias)(c))
}

// marshalJSON handles the JSON encoding with proper escaping
func marshalJSON(data any) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false) // Prevent HTML escaping

	if err := enc.Encode(data); err != nil {
		return nil, err
	}

	b := buf.Bytes()
	if len(b) > 0 && b[len(b)-1] == '\n' {
		b = b[:len(b)-1] // Remove trailing newline if present
	}

	cleanedStr := strings.ReplaceAll(string(b), `"`, `\"`)
	return []byte(cleanedStr), nil
}

// unmarshalJSON is a generic helper for unmarshaling JSON data
func unmarshalJSON[T any](data []byte) (*T, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON data: %w", err)
	}
	return &result, nil
}

// parseTemplateJSON parses the JSON data based on template path
func ParseTemplateFromPath(jsonDataStr string, templatePath TemplatePaths) (any, error) {
	jsonBytes := []byte(strings.ReplaceAll(jsonDataStr, `\`, ""))

	switch templatePath {
	case TemplatePathCountdown, TemplatePathCountdownToTime:
		return unmarshalJSON[Countdown](jsonBytes)
	case TemplatePathDJCountdown, TemplatePathDJCountdownToTime:
		return unmarshalJSON[DJCountdown](jsonBytes)
	case TemplatePathTitle:
		return unmarshalJSON[Title](jsonBytes)
	case TemplatePathSchedule:
		return unmarshalJSON[Schedule](jsonBytes)
	case TemplatePathLowerThird:
		return unmarshalJSON[LowerThird](jsonBytes)
	default:
		return nil, fmt.Errorf("unsupported template path for JSON data: %s", templatePath)
	}
}
