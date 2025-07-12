package types

type Countdown struct {
	Title        string `json:"_title"`
	TimerMinutes string `json:"_timerMinutes"`
	TimerHours   string `json:"_timerHours"`
}

type Title struct {
	Title string `json:"_Title"`
}

type Bar struct {
	Number string `json:"_num"`
	Title  string `json:"_name"`
}

type ScheduleBar struct {
	Row1 string `json:"_row1"`
	Row2 string `json:"_row2"`
	Row3 string `json:"_row3"`

	StartTime string `json:"_timeFrom"`
	EndTime   string `json:"_timeUntil"`

	Hotel string `json:"_hotel"`
	Room  string `json:"_room"`
}

type DetailedDanceComp struct {
	TotalScore string `json:"_num"`
	Name       string `json:"_name"`

	AppearanceScore      string `json:"_numAppearance"`
	ProfessionalismScore string `json:"_numProfessionalism"`
	ConsistencyScore     string `json:"_numConsistency"`
	ComplexityScore      string `json:"_numComplexity"`
	DecibelsScore        string `json:"_numDecibels"`
	OriginalityScore     string `json:"_numOriginality"`
	QuantumScore         string `json:"_numQuantum"`

	PicturePath string `json:"_image"`
	Attribution string `json:"_attribution"`
}

type LowerThird struct {
	Name string `json:"_name"`
	Info string `json:"_info"`
}
