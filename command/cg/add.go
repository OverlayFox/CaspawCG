package cg

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/OverlayFox/CaspawCG/types"
)

type CmdTypeAdd struct {
	CgLayer      int                 `json:"cg_layer"`
	TemplatePath types.TemplatePaths `json:"template_path"`
	PlayOnLoad   bool                `json:"play_on_load"`
	JsonData     []byte              `json:"json_data,omitempty"`
}

func NewCmdTypeAdd(commandParts []string, sheetsHandler types.SheetsHandler, updater types.Updater, layer int) (*CmdTypeAdd, error) {
	if len(commandParts) < 6 {
		return nil, fmt.Errorf("add command must have at least 6 parts")
	}

	cgLayer, err := strconv.Atoi(commandParts[3])
	if err != nil {
		return nil, fmt.Errorf("invalid cg_layer: %s", commandParts[3])
	}

	templatePath, err := types.TemplatePathFromString(commandParts[4])
	if err != nil {
		return nil, fmt.Errorf("invalid template_path: %s", commandParts[4])
	}

	playOnLoad := commandParts[5] == "1"

	var jsonData []byte
	jsonData, err = sheetsHandler.GetData(layer, templatePath)
	if err != nil {
		if len(commandParts) < 7 {
			data, err := types.ParseTemplateFromPath(commandParts[6], templatePath)
			if err != nil {
				return nil, fmt.Errorf("invalid json_data for template '%s': %w", templatePath, err)
			}
			jsonData, err = json.Marshal(data)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal json_data for template '%s': %w", templatePath, err)
			}
		} else {
			return nil, fmt.Errorf("missing json_data for template '%s'", templatePath)
		}
	}

	cmd := &CmdTypeAdd{
		CgLayer:      cgLayer,
		TemplatePath: templatePath,
		PlayOnLoad:   playOnLoad,
		JsonData:     jsonData,
	}

	if layer == 41 && cgLayer == 1 { // if the command is for layer 41 and cgLayer 1 (the schedule), start the updater
		updater.Start(layer) // start the update cycle to send updates for the schedule
	}

	return cmd, nil
}

func (c *CmdTypeAdd) getCmdParts() []string {
	cmdParts := []string{
		strconv.Itoa(c.CgLayer),
		fmt.Sprintf("\"%s\"", string(c.TemplatePath)),
	}

	if c.PlayOnLoad {
		cmdParts = append(cmdParts, "1")
	} else {
		cmdParts = append(cmdParts, "0")
	}

	if len(c.JsonData) > 0 {
		jsonString := string(c.JsonData)
		cmdParts = append(cmdParts, fmt.Sprintf("\"%s\"", jsonString))
	}

	return cmdParts
}
