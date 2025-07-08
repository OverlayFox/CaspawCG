package types

import (
	"fmt"
	"strings"
)

type TemplatePaths string

const (
	TemplatePathCountdown       TemplatePaths = "COUNTDOWN"
	TemplatePathCountdownToTime TemplatePaths = "COUNTDOWN_TO_TIME"
	TemplatePathTitle           TemplatePaths = "TITLE"
	TemplatePathBarBlue         TemplatePaths = "BAR_BLUE"
	TemplatePathBarRed          TemplatePaths = "BAR_RED"
)

var commandTemplatePathMap = map[string]TemplatePaths{
	"COUNTDOWN":         TemplatePathCountdown,
	"COUNTDOWN_TO_TIME": TemplatePathCountdownToTime,
	"TITLE":             TemplatePathTitle,
	"BAR_BLUE":          TemplatePathBarBlue,
	"BAR_RED":           TemplatePathBarRed,
}

func TemplatePathFromString(s string) (TemplatePaths, error) {
	if tp, ok := commandTemplatePathMap[strings.ToUpper(s)]; ok {
		return tp, nil
	}
	return "", fmt.Errorf("unknown template path: %s", s)
}
