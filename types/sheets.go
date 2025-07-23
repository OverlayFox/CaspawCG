package types

type SheetsData interface {
	Start()
	Close()

	GetCountdown() *Countdown
	GetCountdownToTime() *Countdown

	GetLowerThirdSingle() *LowerThird
	GetLowerThirdDuo() (*LowerThird, *LowerThird)

	GetDetailedDanceCompSingle() *DetailedDanceComp
	GetDetailedDanceComp() []*DetailedDanceComp

	GetCurrentSchedule() []*ScheduleRow

	GetAttribution(contestantsName string) (string, error)

	GetFreeStandings() []*FreeStandings
}
