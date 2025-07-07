package types

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
