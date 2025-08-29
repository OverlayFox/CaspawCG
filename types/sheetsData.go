package types

import (
	"time"
)

type CountdownType string

const (
	CountdownTypeDuration CountdownType = "duration"
	CountdownTypeToTime   CountdownType = "toTime"
)

type SimpleCountdown struct {
	CountdownType CountdownType `json:"countdownType"`
	Text          string        `json:"text"`
	CountdownTime string        `json:"countdownTime"`
}

type DJCountdown struct {
	CountdownType CountdownType `json:"countdownType"`
	Name          string        `json:"name"`
	Genre         string        `json:"genre"`
	CountdownTime string        `json:"countdownTime"`
}

type LowerThird struct {
	Row1 string `json:"row1"`
	Row2 string `json:"row2"`
}

type ScheduleRow struct {
	Name      string    `json:"name"`
	Genre     string    `json:"genre"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"duration"`
}
