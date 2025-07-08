package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Countdown struct {
	Title         string `json:"title"`
	CountdownTime string `json:"countdownTime"`
}

type CountdownToTime struct {
	Title           string `json:"title"`
	CountdownToTime string `json:"countdownToTime"`
}

type LowerThird struct {
	Row1 string `json:"row1"`
	Row2 string `json:"row2"`
}

type DetailedDanceComp struct {
	Name string `json:"name"`

	Appearance      string `json:"apperance"`
	Professionalism string `json:"professionalism"`
	Consistency     string `json:"consistency"`
	Complexity      string `json:"complexity"`
	Decibels        string `json:"decibles"`
	Originality     string `json:"originality"`
	Quantum         string `json:"quantum"`

	TotalScore string `json:"totalScore"`
}

type ScheduleRow struct {
	Title     string    `json:"title"`
	Hotel     string    `json:"hotel"`
	Room      string    `json:"room"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"duration"`
}

func NewScheduleRow(title, room, weekDay, startTime, durationStr string) (*ScheduleRow, error) {
	loc, err := time.LoadLocation("Europe/Vienna")
	if err != nil {
		return nil, fmt.Errorf("failed to load location: %w", err)
	}

	var day int = 0
	switch strings.ToUpper(strings.TrimSpace(weekDay)) {
	case "THU":
		day = 23
	case "FRI":
		day = 24
	case "SAT":
		day = 25
	case "SUN":
		day = 26
	default:
		return nil, fmt.Errorf("invalid weekday: %s", weekDay)
	}
	startDate := time.Date(2025, time.July, day, 0, 0, 0, 0, loc)

	var second, minute, hour int = 0, 0, 0
	parts := strings.Split(startTime, ":")
	switch len(parts) {
	case 3:
		second, err = strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid second in startTime: %w", err)
		}
		fallthrough
	case 2:
		minute, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid minute in startTime: %w", err)
		}
		hour, err = strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid hour in startTime: %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid startTime format: %s", startTime)
	}
	startDate = startDate.Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute + time.Duration(second)*time.Second)

	durationInt, err := strconv.Atoi(durationStr)
	if err != nil {
		return nil, fmt.Errorf("invalid duration: %w", err)
	}
	duration := time.Duration(durationInt) * time.Minute

	formattedHotel, formattedRoom := getRoom(strings.ToUpper(strings.TrimSpace(room)))

	return &ScheduleRow{
		Title: strings.TrimSpace(title),
		Hotel: formattedHotel,
		Room:  formattedRoom,

		StartTime: startDate,
		EndTime:   startDate.Add(duration),
	}, nil
}

type RoomInfo struct {
	Hotel string
	Room  string
}

func getRoom(room string) (string, string) {
	roomMap := map[string]RoomInfo{
		// Wimberger hotel rooms
		"WIMB_LOBBY":    {"Wimberger", "Lobby"},
		"WIMB_MAIN":     {"Wimberger", "Main Stage"},
		"WIMB_JOESBAR":  {"Wimberger", "Joe's Bar"},
		"WIMB_+1_PANEL": {"Wimberger", "Panel Room +1"},

		// Flemings hotel rooms
		"FLEM_-2_COMMU":   {"Flemings", "Community Room -2"},
		"FLEM_-2_PANEL":   {"Flemings", "Panel Room -2"},
		"FLEM_-1_ARTISTA": {"Flemings", "Artist Ally -1"},
		"FLEM_-1_ARTSHOW": {"Flemings", "Art Show -1"},
		"FLEM_-1_DEALERS": {"Flemings", "Dealers Den -1"},
		"FLEM_-1_PANEL":   {"Flemings", "Panel Room -1"},
		"FLEM_00_BAR":     {"Flemings", "Bar"},
		"FLEM_00_OUTSIDE": {"Flemings", "Outside"},
	}

	for key, info := range roomMap {
		parts := strings.Split(key, "_")
		allMatch := true

		for _, part := range parts {
			if !strings.Contains(room, part) {
				allMatch = false
				break
			}
		}

		if allMatch {
			return info.Hotel, info.Room
		}
	}

	return "", ""
}
