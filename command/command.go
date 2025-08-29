package command

import (
	"fmt"

	"github.com/OverlayFox/CaspawCG/command/cg"
	"github.com/OverlayFox/CaspawCG/types"
)

type Handler struct {
	CommandType types.CommandType `json:"command_type"`
	Command     any               `json:"-"`
}

func (h *Handler) GetCommandString() (string, error) {
	if h.Command == nil {
		return "", fmt.Errorf("command is not set")
	}

	switch h.CommandType {
	case types.CommandTypeCG:
		cmd, ok := h.Command.(*cg.CmdCG)
		if !ok {
			return "", fmt.Errorf("invalid command type: %s", h.CommandType)
		}
		return cmd.GetCommandString()
	}

	return "", fmt.Errorf("unsupported command type: %s", h.CommandType)
}

// NewHandlerFromCmdString return a Command Handler that was build from an incoming command string
func NewHandlerFromCmdString(cmd string, sheetsHandler types.SheetsHandler, updater types.Updater) (types.CommandHandler, error) {
	parts, err := ParseCommandLine(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to parse command line: %w", err)
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("command cannot be empty")
	}

	commandType, err := types.CommandTypeFromString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("%w: %s", types.ErrUnsupportedCommandType, parts[0])
	}

	if commandType != types.CommandTypeCG {
		return nil, fmt.Errorf("%w: '%s'", types.ErrUnsupportedCommandType, commandType)
	}

	command, err := cg.NewCommandCG(parts, sheetsHandler, updater)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", types.ErrUnsupportedCommandType, err)
	}

	return &Handler{
		CommandType: commandType,
		Command:     command,
	}, nil
}
