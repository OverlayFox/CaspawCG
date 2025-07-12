package types

import (
	"fmt"
	"strings"
)

type CommandCall string

const (
	CommandCallADD  CommandCall = "ADD"
	CommandCallSTOP CommandCall = "STOP"
)

var commandCallMap = map[string]CommandCall{
	"ADD":  CommandCallADD,
	"STOP": CommandCallSTOP,
}

func CommandCallFromString(s string) (CommandCall, error) {
	if ct, ok := commandCallMap[strings.ToUpper(s)]; ok {
		return ct, nil
	}
	return "", fmt.Errorf("unknown command type: %s", s)
}
