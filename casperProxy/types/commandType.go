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
	CommandTypePlay    CommandType = "PLAY"
	CommandTypeStop    CommandType = "STOP"
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
	"PLAY":      CommandTypePlay,
	"STOP":      CommandTypeStop,
}

func CommandTypeFromString(s string) (CommandType, error) {
	if ct, ok := commandTypeMap[strings.ToUpper(s)]; ok {
		return ct, nil
	}
	return "", fmt.Errorf("unknown command type: %s", s)
}
