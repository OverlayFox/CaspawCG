package types

import (
	"time"
)

type CountdownType string

const (
	CountdownTypeDuration CountdownType = "duration"
	CountdownTypeToTime   CountdownType = "toTime"
)

type SheetsCountdown struct {
	CountdownType CountdownType `json:"countdownType"`
	Text          string        `json:"text"`
	CountdownTime string        `json:"countdownTime"`
}

type SheetsDJCountdown struct {
	CountdownType CountdownType `json:"countdownType"`
	Name          string        `json:"name"`
	Genre         string        `json:"genre"`
	CountdownTime string        `json:"countdownTime"`
}

type SheetsLowerThird struct {
	Row1 string `json:"row1"`
	Row2 string `json:"row2"`
}

type SheetsScheduleRow struct {
	Name      string    `json:"name"`
	Genre     string    `json:"genre"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"duration"`
}
