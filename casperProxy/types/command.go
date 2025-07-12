package types

import (
	"fmt"
)

type Command interface {
	GetCommandType() CommandType
}

type CommandStruct struct {
	CommandType CommandType `json:"command_type"`
}

func (cs CommandStruct) GetCommandType() CommandType {
	return cs.CommandType
}

func NewCommandFromString(command string) (Command, error) {
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
