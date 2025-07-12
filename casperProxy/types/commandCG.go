package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// CommandCG represents a CG command with all its parameters
type CommandCG struct {
	CommandStruct
	Channel      *int           `json:"channel,omitempty"`
	Layer        *int           `json:"layer,omitempty"`
	Call         *CommandCall   `json:"call,omitempty"`
	CgLayer      *int           `json:"cg_layer,omitempty"`
	TemplatePath *TemplatePaths `json:"template_path,omitempty"`
	PlayOnLoad   *bool          `json:"play_on_load,omitempty"`
	JsonData     any            `json:"json_data,omitempty"`
}

// Command generates the command string from the CommandCG struct
func (c *CommandCG) Command() (string, error) {
	if c.TemplatePath == nil {
		return "", fmt.Errorf("template path must be specified")
	}

	commandParts := []string{
		string(c.CommandType),
		fmt.Sprintf("%d-%d", *c.Channel, *c.Layer),
		string(*c.Call),
		strconv.Itoa(*c.CgLayer),
		fmt.Sprintf("\"%s\"", string(*c.TemplatePath)),
	}

	if c.PlayOnLoad != nil {
		playOnLoadValue := "0"
		if *c.PlayOnLoad {
			playOnLoadValue = "1"
		}
		commandParts = append(commandParts, playOnLoadValue)
	}

	if c.JsonData != nil {
		jsonString, err := c.encodeJSONData()
		if err != nil {
			return "", fmt.Errorf("failed to encode JSON data: %w", err)
		}
		commandParts = append(commandParts, fmt.Sprintf("\"%s\"", jsonString))
	}

	return strings.Join(commandParts, " "), nil
}

// encodeJSONData handles the JSON encoding with proper escaping
func (c *CommandCG) encodeJSONData() (string, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false) // Prevent HTML escaping

	if err := enc.Encode(c.JsonData); err != nil {
		return "", err
	}

	b := buf.Bytes()
	if len(b) > 0 && b[len(b)-1] == '\n' {
		b = b[:len(b)-1] // Remove trailing newline if present
	}

	return strings.ReplaceAll(string(b), `"`, `\"`), nil
}

// NewCommandCG creates a new CommandCG from command parts
func NewCommandCG(commandParts []string) (*CommandCG, error) {
	if len(commandParts) < 3 {
		return nil, fmt.Errorf("command must have at least 3 parts")
	}

	cmd := &CommandCG{
		CommandStruct: CommandStruct{
			CommandType: CommandTypeCG,
		},
	}

	if err := cmd.parseChannelLayer(commandParts[1]); err != nil {
		return nil, err
	}

	if err := cmd.parseCommandCall(commandParts[2]); err != nil {
		return nil, err
	}

	if len(commandParts) >= 4 {
		if err := cmd.parseCgLayer(commandParts[3]); err != nil {
			return nil, err
		}
	}

	if len(commandParts) >= 5 {
		if err := cmd.parseTemplatePath(commandParts[4]); err != nil {
			return nil, err
		}
	}

	if len(commandParts) >= 6 {
		if err := cmd.parsePlayOnLoad(commandParts[5]); err != nil {
			return nil, err
		}
	}

	if len(commandParts) >= 7 {
		if err := cmd.parseJSONData(commandParts[6]); err != nil {
			return nil, err
		}
	}

	return cmd, nil
}

// parseChannelLayer parses the channel-layer format
func (c *CommandCG) parseChannelLayer(channelLayerStr string) error {
	parts := strings.Split(channelLayerStr, "-")
	if len(parts) != 2 {
		return fmt.Errorf("invalid channel-layer format: %s", channelLayerStr)
	}

	channel, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid channel: %s", parts[0])
	}
	c.Channel = &channel

	layer, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid layer: %s", parts[1])
	}
	c.Layer = &layer

	return nil
}

// parseCommandCall parses the command call
func (c *CommandCG) parseCommandCall(callStr string) error {
	if callStr == "" {
		return fmt.Errorf("command call cannot be empty")
	}

	commandCall, err := CommandCallFromString(callStr)
	if err != nil {
		return fmt.Errorf("invalid command call: %s", callStr)
	}
	c.Call = &commandCall

	return nil
}

// parseCgLayer parses the CG layer
func (c *CommandCG) parseCgLayer(layerStr string) error {
	if layerStr == "" {
		return fmt.Errorf("cg_layer cannot be empty")
	}

	cgLayer, err := strconv.Atoi(layerStr)
	if err != nil {
		return fmt.Errorf("invalid cg_layer: %s", layerStr)
	}
	c.CgLayer = &cgLayer

	return nil
}

// parseTemplatePath parses the template path
func (c *CommandCG) parseTemplatePath(pathStr string) error {
	if pathStr == "" {
		return fmt.Errorf("template path cannot be empty")
	}

	templatePath, err := TemplatePathFromString(pathStr)
	if err != nil {
		return fmt.Errorf("invalid template path: %s", pathStr)
	}
	c.TemplatePath = &templatePath

	return nil
}

// parsePlayOnLoad parses the play on load flag
func (c *CommandCG) parsePlayOnLoad(playOnLoadStr string) error {
	switch playOnLoadStr {
	case "1":
		c.PlayOnLoad = new(bool)
		*c.PlayOnLoad = true
	case "0":
		c.PlayOnLoad = new(bool)
		*c.PlayOnLoad = false
	default:
		return fmt.Errorf("invalid play_on_load value: %s", playOnLoadStr)
	}

	return nil
}

// parseJSONData parses the JSON data based on template path
func (c *CommandCG) parseJSONData(jsonDataStr string) error {
	if c.TemplatePath == nil {
		return fmt.Errorf("template path must be specified to unmarshal JSON data")
	}

	jsonBytes := []byte(strings.ReplaceAll(jsonDataStr, `\`, ""))

	var err error
	switch *c.TemplatePath {
	case TemplatePathCountdown, TemplatePathCountdownToTime:
		c.JsonData, err = unmarshalJSON[Countdown](jsonBytes)
	case TemplatePathTitle:
		c.JsonData, err = unmarshalJSON[Title](jsonBytes)
	case TemplatePathBarRed, TemplatePathBarBlue:
		c.JsonData, err = unmarshalJSON[Bar](jsonBytes)
	case TemplatePathSchedule:
		c.JsonData, err = unmarshalJSON[ScheduleBar](jsonBytes)
	case TemplatePathDanceComp:
		c.JsonData, err = unmarshalJSON[DetailedDanceComp](jsonBytes)
	case TemplatePathLowerThird:
		c.JsonData, err = unmarshalJSON[LowerThird](jsonBytes)
	default:
		return fmt.Errorf("unsupported template path for JSON data: %s", *c.TemplatePath)
	}

	return err
}

// unmarshalJSON is a generic helper for unmarshaling JSON data
func unmarshalJSON[T any](data []byte) (*T, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON data: %w", err)
	}
	return &result, nil
}
