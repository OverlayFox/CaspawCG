package types

import (
	"fmt"
	"strings"
)

type CommandType string

const (
	CommandTypeCG      CommandType = "CG"
	CommandTypeINFO    CommandType = "INFO"
	CommandTypeCLS     CommandType = "CLS"
	CommandTypeTS      CommandType = "TLS"
	CommandTypeDATA    CommandType = "DATA"
	CommandTypeVERSION CommandType = "VERSION"
	CommandTypeUNKNOWN CommandType = "THUMBNAIL"
	CommandTypeMIXER   CommandType = "MIXER"
	CommandTypeFILL    CommandType = "FILL"
	CommandTypeClear   CommandType = "CLEAR"
)

var commandTypeMap = map[string]CommandType{
	"CG":        CommandTypeCG,
	"INFO":      CommandTypeINFO,
	"CLS":       CommandTypeCLS,
	"TLS":       CommandTypeTS,
	"DATA":      CommandTypeDATA,
	"VERSION":   CommandTypeVERSION,
	"THUMBNAIL": CommandTypeUNKNOWN,
	"MIXER":     CommandTypeMIXER,
	"FILL":      CommandTypeFILL,
	"CLEAR":     CommandTypeClear,
}

func CommandTypeFromString(s string) (CommandType, error) {
	if ct, ok := commandTypeMap[strings.ToUpper(s)]; ok {
		return ct, nil
	}
	return "", fmt.Errorf("unknown command type: %s", s)
}

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

type CommandStruct struct {
	CommandType CommandType `json:"command_type"`
}

func (cs CommandStruct) GetCommandType() CommandType {
	return cs.CommandType
}

type Command interface {
	GetCommandType() CommandType
}

func NewCommand(command string) (Command, error) {
	parts, err := ParseCommandLine(command)
	if err != nil {
		return nil, fmt.Errorf("failed to parse command line: %w", err)
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("command cannot be empty")
	}

	commandType, err := CommandTypeFromString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid command type: %s", parts[0])
	}

	switch commandType {
	case CommandTypeCG:
		return NewCommandCG(parts)
	}

	return nil, fmt.Errorf("%w: %s", ErrUnsupportedCommandType, commandType)
}
