package types

import (
	"fmt"
	"strings"
)

type TemplatePaths string

const (
	TemplatePathCountdown TemplatePaths = "COUNTDOWN"
	TemplatePathTitle     TemplatePaths = "TITLE"
	TemplatePathBar       TemplatePaths = "BAR"
)

var commandTemplatePathMap = map[string]TemplatePaths{
	"COUNTDOWN":         TemplatePathCountdown,
	"COUNTDOWN_TO_TIME": TemplatePathCountdown,
	"TITLE":             TemplatePathTitle,
	"BAR_BLUE":          TemplatePathBar,
	"BAR_RED":           TemplatePathBar,
}

func TemplatePathFromString(s string) (TemplatePaths, error) {
	if tp, ok := commandTemplatePathMap[strings.ToUpper(s)]; ok {
		return tp, nil
	}
	return "", fmt.Errorf("unknown template path: %s", s)
}
