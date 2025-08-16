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
	*SimpleCountdown
	FirstDJ  string `json:"firstDJ"`
	SecondDJ string `json:"secondDJ"`
	Genre    string `json:"genre"`
}

type LowerThird struct {
	Row1 string `json:"row1"`
	Row2 string `json:"row2"`
}

type ScheduleRow struct {
	Title     string    `json:"title"`
	Genre     string    `json:"genre"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"duration"`
}
