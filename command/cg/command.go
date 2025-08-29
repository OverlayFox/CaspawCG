package cg

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/OverlayFox/CaspawCG/types"
)

// CmdCG represents a CG command with all its parameters
type CmdCG struct {
	Channel     int                   `json:"channel"`
	Layer       int                   `json:"layer"`
	Call        types.CommandCallType `json:"call"`
	CommandCall any                   `json:"-"` // use the call type to type cast CommandCall

	ctx    context.Context
	cancel context.CancelFunc
}

// NewCommandCG creates a new CommandCG from command parts
func NewCommandCG(commandParts []string, sheetsHandler types.SheetsHandler, updater types.Updater) (*CmdCG, error) {
	if len(commandParts) < 3 {
		return nil, fmt.Errorf("command must have at least 3 parts")
	}

	cmd := &CmdCG{}

	if err := cmd.parseChannelAndLayer(commandParts[1]); err != nil {
		return nil, err
	}

	if err := cmd.parseCommandCall(commandParts[2]); err != nil {
		return nil, err
	}

	switch cmd.Call {
	case types.CommandCallTypeADD:
		cmdAdd, err := NewCmdTypeAdd(commandParts, sheetsHandler, updater, cmd.Layer)
		if err != nil {
			return nil, err
		}
		cmd.CommandCall = cmdAdd

	case types.CommandCallTypeUPDATE:
		cmdUpdate, err := NewCmdTypeUpdate(commandParts, sheetsHandler, updater, cmd.Layer)
		if err != nil {
			return nil, err
		}
		cmd.CommandCall = cmdUpdate

	case types.CommandCallTypeSTOP:
		if cmd.Layer == 41 {
			updater.Stop() // stop the update cycle when STOP is called on layer 41
		}
		return nil, fmt.Errorf("STOP command does not require additional parameters")

	default:
		return nil, fmt.Errorf("unsupported command call: %s", cmd.Call)
	}

	return cmd, nil
}

// GetCommandString generates the command string from the CmdCG struct
func (c *CmdCG) GetCommandString() (string, error) {
	cmdParts := []string{
		string(types.CommandTypeCG),
		fmt.Sprintf("%d-%d", c.Channel, c.Layer),
		string(c.Call),
	}

	var extraParts []string
	var err error
	switch c.Call {
	case types.CommandCallTypeADD:
		extraParts, err = c.buildAddCommand()
	case types.CommandCallTypeUPDATE:
		extraParts, err = c.buildUpdateCommand()
	}

	if err != nil {
		return "", err
	}

	return strings.Join(append(cmdParts, extraParts...), " "), nil
}

func (c *CmdCG) buildAddCommand() ([]string, error) {

	cmdAddParts, ok := c.CommandCall.(*CmdTypeAdd)
	if !ok {
		return nil, fmt.Errorf("invalid command call for ADD: %T", c.CommandCall)
	}

	return cmdAddParts.getCmdParts(), nil
}

func (c *CmdCG) buildUpdateCommand() ([]string, error) {
	cmdUpdateParts, ok := c.CommandCall.(*CmdTypeUpdate)
	if !ok {
		return nil, fmt.Errorf("invalid command call for UPDATE: %T", c.CommandCall)
	}

	return cmdUpdateParts.getCmdParts(), nil
}

// parseChannelLayer parses the channel-layer format
func (c *CmdCG) parseChannelAndLayer(channelLayerStr string) error {
	parts := strings.Split(channelLayerStr, "-")
	if len(parts) != 2 {
		return fmt.Errorf("invalid channel-layer format: %s", channelLayerStr)
	}

	channel, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid channel: %s", parts[0])
	}
	c.Channel = channel

	layer, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid layer: %s", parts[1])
	}
	c.Layer = layer

	return nil
}

// parseCommandCall parses the command call
func (c *CmdCG) parseCommandCall(callStr string) error {
	if callStr == "" {
		return fmt.Errorf("command call cannot be empty")
	}

	commandCall, err := types.CommandCallFromString(callStr)
	if err != nil {
		return fmt.Errorf("invalid command call: %s", callStr)
	}
	c.Call = commandCall

	return nil
}
