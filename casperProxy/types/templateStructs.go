package types

type Countdown struct {
	Title        string `json:"_title"`
	TimerMinutes string `json:"_timerMinutes"`
	TimerHours   string `json:"_timerHours"`
}

type DJCountdown struct {
	Name         string `json:"_name"`
	Genre        string `json:"_genre"`
	Logo         string `json:"_logo"`
	TimerMinutes string `json:"_timerMinutes"`
	TimerHours   string `json:"_timerHours"`
}

type Title struct {
	Title string `json:"_title"`
}

type Schedule struct {
	Name01  string `json:"_name01"`
	Genre01 string `json:"_genre01"`
	Time01  string `json:"_time01"`
	Logo01  string `json:"_logo01"`

	Name02  string `json:"_name02"`
	Genre02 string `json:"_genre02"`
	Time02  string `json:"_time02"`
	Logo02  string `json:"_logo02"`

	Name03  string `json:"_name03"`
	Genre03 string `json:"_genre03"`
	Time03  string `json:"_time03"`
	Logo03  string `json:"_logo03"`
}

type LowerThird struct {
	Name string `json:"_name"`
	Info string `json:"_info"`
}
