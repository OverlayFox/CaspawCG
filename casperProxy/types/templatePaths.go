package types

import (
	"fmt"
	"strings"
)

type TemplatePaths string

const (
	TemplatePathCountdown TemplatePaths = "COUNTDOWN"
)

var commandTemplatePathMap = map[string]TemplatePaths{
	"COUNTDOWN":         TemplatePathCountdown,
	"COUNTDOWN_TO_TIME": TemplatePathCountdown,
}

func TemplatePathFromString(s string) (TemplatePaths, error) {
	if tp, ok := commandTemplatePathMap[strings.ToUpper(s)]; ok {
		return tp, nil
	}
	return "", fmt.Errorf("unknown template path: %s", s)
}
