package ui

import (
	"encoding/json"
	"os"
)

type FieldConfig struct {
	Key    string `json:"key"`
	Type   string `json:"type"`
	ID     string `json:"id"`
	Source string `json:"source"`
}

type WidgetConfig struct {
	ID       string        `json:"id"`
	X        int           `json:"x"`
	Y        int           `json:"y"`
	W        int           `json:"w"`
	H        int           `json:"h"`
	Template string        `json:"template"`
	Layer    int           `json:"layer"`
	Channel  int           `json:"channel"`
	PosX     *int          `json:"posX,omitempty"`
	PosY     *int          `json:"posY,omitempty"`
	SizeX    *float64      `json:"sizeX,omitempty"`
	SizeY    *float64      `json:"sizeY,omitempty"`
	Delay    int           `json:"delay,omitempty"`
	Fields   []FieldConfig `json:"fields"`
}

type MediaWidgetConfig struct {
	ID       string `json:"id"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	W        int    `json:"w"`
	H        int    `json:"h"`
	Filename string `json:"filename"`
	Layer    int    `json:"layer"`
	Channel  int    `json:"channel"`
	Delay    int    `json:"delay,omitempty"`
	Loop     bool   `json:"loop"`
}

type GroupConfig struct {
	ID      string         `json:"id"`
	X       int            `json:"x"`
	Y       int            `json:"y"`
	W       int            `json:"w"`
	H       int            `json:"h"`
	Name    string         `json:"name"`
	Widgets []WidgetConfig `json:"widgets"`
}

type LayoutConfig struct {
	Version      int                 `json:"version"`
	Widgets      []WidgetConfig      `json:"widgets"`
	Groups       []GroupConfig       `json:"groups,omitempty"`
	MediaWidgets []MediaWidgetConfig `json:"mediaWidgets,omitempty"`
}

const layoutFileName = "layout.json"

func SaveLayout(config LayoutConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(layoutFileName, data, 0o644)
}

func LoadLayout() (LayoutConfig, error) {
	var config LayoutConfig

	data, err := os.ReadFile(layoutFileName)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return LayoutConfig{Version: 1, Widgets: []WidgetConfig{}}, nil
		}
		return config, err
	}

	err = json.Unmarshal(data, &config)
	return config, err
}
