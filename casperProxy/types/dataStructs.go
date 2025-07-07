package types

import (
	"encoding/json"
	"net/url"
)

type Countdown struct {
	Title        string `json:"_title"`
	TimerMinutes string `json:"_timerMinutes"`
	TimerHours   string `json:"_timerHours"`
}

func (c *Countdown) Command() (string, error) {
	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	escaped := url.QueryEscape(string(jsonBytes))
	return escaped, nil
}

type Title struct {
	Title string `json:"_Title"`
}

func (t *Title) Command() (string, error) {
	jsonBytes, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	escaped := url.QueryEscape(string(jsonBytes))
	return escaped, nil
}

type Bar struct {
	Number string `json:"_num"`
	Title  string `json:"_name"`
}

func (b *Bar) Command() (string, error) {
	jsonBytes, err := json.Marshal(b)
	if err != nil {
		return "", err
	}
	escaped := url.QueryEscape(string(jsonBytes))
	return escaped, nil
}
