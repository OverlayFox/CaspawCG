package ui

import (
	"encoding/json"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/overlayfox/caspaw-cg/src/types"
)

type FieldConfig struct {
	Key       string         `json:"key"`
	Type      types.DataType `json:"type"`
	InputType string         `json:"inputType"`
	Location  string         `json:"location,omitempty"`
	Source    string         `json:"source,omitempty"`
	Value     string         `json:"value,omitempty"`
	Range     string         `json:"range,omitempty"`
	Offset    int            `json:"offset,omitempty"`
}

type TemplateConfig struct {
	ID               string       `json:"id"`
	X                int          `json:"x"`
	Y                int          `json:"y"`
	W                int          `json:"w"`
	H                int          `json:"h"`
	Name             string       `json:"name,omitempty"`
	Template         string       `json:"template"`
	ChannelExpr      string       `json:"channelExpr,omitempty"`
	Layer            int          `json:"layer"`
	Sizing           types.Sizing `json:"sizing"`
	DelayMs          int          `json:"delayMs,omitempty"`
	UpdateIntervalMs int          `json:"updateIntervalMs,omitempty"`

	ScheduleStartTimeColumn string `json:"scheduleStartTimeColumn,omitempty"`
	ScheduleEndTimeColumn   string `json:"scheduleEndTimeColumn,omitempty"`

	Fields []FieldConfig `json:"fields"`
}

type MediaWidgetConfig struct {
	ID          string `json:"id"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	W           int    `json:"w"`
	H           int    `json:"h"`
	Name        string `json:"name,omitempty"`
	Filename    string `json:"filename"`
	Layer       int    `json:"layer"`
	ChannelExpr string `json:"channelExpr,omitempty"`
	DelayMs     int    `json:"delayMs,omitempty"`
	Loop        bool   `json:"loop"`
}

type GroupConfig struct {
	ID           string              `json:"id"`
	X            int                 `json:"x"`
	Y            int                 `json:"y"`
	W            int                 `json:"w"`
	H            int                 `json:"h"`
	Name         string              `json:"name"`
	Widgets      []TemplateConfig    `json:"widgets"`
	MediaWidgets []MediaWidgetConfig `json:"mediaWidgets,omitempty"`
}

type LayoutConfig struct {
	Version      int                 `json:"version"`
	Widgets      []TemplateConfig    `json:"widgets"`
	Groups       []GroupConfig       `json:"groups,omitempty"`
	MediaWidgets []MediaWidgetConfig `json:"mediaWidgets,omitempty"`
}

const layoutVersion = 2

const layoutFileName = "layout.json"

func emptyLayout() LayoutConfig {
	return LayoutConfig{Version: layoutVersion, Widgets: []TemplateConfig{}}
}

func SaveLayout(config LayoutConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(layoutFileName, data, 0o644)
}

// LoadLayout reads and parses layout.json. A missing file or a file that fails to
// parse against the current schema (e.g. one saved by an older, incompatible version
// of the app) is treated as "no layout yet" rather than an error, so a schema change
// self-heals on first run instead of crashing the app.
func LoadLayout() (LayoutConfig, error) {
	data, err := os.ReadFile(layoutFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return emptyLayout(), nil
		}
		return emptyLayout(), err
	}

	var config LayoutConfig
	if err := json.Unmarshal(data, &config); err != nil {
		log.Warn().Err(err).Msg("Failed to parse layout.json against the current schema; starting with an empty layout")
		return emptyLayout(), nil
	}
	return config, nil
}
