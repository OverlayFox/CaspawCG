package types

import (
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
