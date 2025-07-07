package types

type Countdown struct {
	Title        string `json:"_title"`
	TimerMinutes string `json:"_timerMinutes"`
	TimerSeconds string `json:"_timerHours"`
}

type Title struct {
	Title string `json:"_Title"`
}

type Bar struct {
	Number string `json:"_num"`
	Title  string `json:"_name"`
}
