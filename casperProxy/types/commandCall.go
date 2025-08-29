package types

import (
	"fmt"
	"strings"
)

type CommandCall string

const (
	CommandCallADD    CommandCall = "ADD"
	CommandCallSTOP   CommandCall = "STOP"
	CommandCallUPDATE CommandCall = "UPDATE"
)

var commandCallMap = map[string]CommandCall{
	"ADD":    CommandCallADD,
	"STOP":   CommandCallSTOP,
	"UPDATE": CommandCallUPDATE,
}

func CommandCallFromString(s string) (CommandCall, error) {
	if ct, ok := commandCallMap[strings.ToUpper(s)]; ok {
		return ct, nil
	}
	return "", fmt.Errorf("unknown command type: %s", s)
}
