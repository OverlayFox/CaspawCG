package cg

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/OverlayFox/CaspawCG/types"
)

type CmdTypeUpdate struct {
	CgLayer      int                 `json:"cg_layer"`
	TemplatePath types.TemplatePaths `json:"-"`
	JsonData     []byte              `json:"json_data,omitempty"`
}

func NewCmdTypeUpdate(commandParts []string, sheetsHandler types.SheetsHandler, updater types.Updater, layer int) (*CmdTypeUpdate, error) {
	if len(commandParts) < 4 {
		return nil, fmt.Errorf("update command must have at least 4 parts")
	}

	cgLayer, err := strconv.Atoi(commandParts[3])
	if err != nil {
		return nil, fmt.Errorf("invalid cg_layer: %s", commandParts[3])
	}

	templatePath, err := types.TemplateFromLayer(cgLayer)
	if err != nil {
		return nil, err
	}

	var jsonData []byte
	if len(commandParts) >= 5 {
		data, err := types.ParseTemplateFromPath(commandParts[4], templatePath)
		if err != nil {
			return nil, err
		}
		jsonData, err = json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal json_data for template '%s': %w", templatePath, err)
		}
	}

	return &CmdTypeUpdate{
		CgLayer:      cgLayer,
		TemplatePath: templatePath,
		JsonData:     jsonData,
	}, nil
}

func (c *CmdTypeUpdate) getCmdParts() []string {
	cmdParts := []string{
		strconv.Itoa(c.CgLayer),
	}

	if len(c.JsonData) > 0 {
		jsonString := string(c.JsonData)
		cmdParts = append(cmdParts, fmt.Sprintf("\"%s\"", jsonString))
	}

	return cmdParts
}
