package types

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

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

func NewCommandCG(commandParts []string) (*CommandCG, error) {
	if len(commandParts) < 3 {
		return nil, fmt.Errorf("command must have at least 3 parts")
	}

	var (
		channel, layer *int
		commandCall    *CommandCall
		cgLayer        *int
		templatePath   *TemplatePaths
		playOnLoad     *bool
		jsonDataParsed any
	)

	switch len(commandParts) {
	case 7: // JSON Data
		fallthrough
	case 6: // PlayOnLoad
		if commandParts[5] == "1" {
			playOnLoad = new(bool)
			*playOnLoad = true
		} else if commandParts[5] == "0" {
			playOnLoad = new(bool)
			*playOnLoad = false
		} else {
			return nil, fmt.Errorf("invalid play_on_load value: %s", commandParts[5])
		}
		fallthrough
	case 5: // Template Path
		fmt.Println("Processing Template Path:", commandParts[4])
		if commandParts[4] == "" {
			return nil, fmt.Errorf("template path cannot be empty")
		}
		templatePathType, err := TemplatePathFromString(commandParts[4])
		if err != nil {
			return nil, fmt.Errorf("invalid template path: %s", commandParts[4])
		}
		templatePath = &templatePathType
		fallthrough
	case 4: // CG Layer
		if commandParts[3] == "" {
			return nil, fmt.Errorf("cg_layer cannot be empty")
		}
		cgLayerInt, err := strconv.Atoi(commandParts[3])
		if err != nil {
			return nil, fmt.Errorf("invalid cg_layer: %s", commandParts[3])
		}
		cgLayer = new(int)
		*cgLayer = cgLayerInt
		fallthrough
	case 3: // Command Call
		if commandParts[2] == "" {
			return nil, fmt.Errorf("command call cannot be empty")
		}
		commandCallStr, err := CommandCallFromString(commandParts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid command call: %s", commandParts[2])
		}
		commandCall = &commandCallStr
		fallthrough
	case 2: // Channel and Layer
		channelLayer := strings.Split(commandParts[1], "-")
		if len(channelLayer) != 2 {
			return nil, fmt.Errorf("invalid channel-layer format: %s", commandParts[1])
		}
		channelInt, err := strconv.Atoi(channelLayer[0])
		if err != nil {
			return nil, fmt.Errorf("invalid channel: %s", channelLayer[0])
		}
		channel = new(int)
		*channel = channelInt

		layerInt, err := strconv.Atoi(channelLayer[1])
		if err != nil {
			return nil, fmt.Errorf("invalid layer: %s", channelLayer[1])
		}
		layer = new(int)
		*layer = layerInt
	}

	// JSON Data
	if len(commandParts) == 7 {
		jsonBytes := []byte(strings.ReplaceAll(commandParts[6], `\`, ""))
		if templatePath == nil {
			return nil, fmt.Errorf("template path must be specified to unmarshal JSON data")
		}

		switch *templatePath {
		case TemplatePathCountdown:
			var countdownData Countdown
			if err := json.Unmarshal(jsonBytes, &countdownData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal countdown data: %w", err)
			}
			jsonDataParsed = &countdownData
		case TemplatePathTitle:
			var titleData Title
			if err := json.Unmarshal(jsonBytes, &titleData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal title data: %w", err)
			}
			jsonDataParsed = &titleData
		case TemplatePathBar:
			var barData Bar
			if err := json.Unmarshal(jsonBytes, &barData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal bar data: %w", err)
			}
			jsonDataParsed = &barData
		default:
			return nil, fmt.Errorf("unsupported template path for JSON data: %s", *templatePath)
		}
	}

	return &CommandCG{
		CommandStruct: CommandStruct{
			CommandType: CommandTypeCG,
		},
		Channel:      channel,
		Layer:        layer,
		Call:         commandCall,
		CgLayer:      cgLayer,
		TemplatePath: templatePath,
		PlayOnLoad:   playOnLoad,
		JsonData:     jsonDataParsed,
	}, nil
}
