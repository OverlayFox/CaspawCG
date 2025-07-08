package types

type SheetsData interface {
	Start()

	GetCountdown() *Countdown
	GetCountdownToTime() *Countdown

	GetLowerThird() *LowerThird

	GetDetailedDanceCompSingle() *DetailedDanceComp
	GetDetailedDanceComp() []*DetailedDanceComp

	GetCurrentSchedule() []*ScheduleRow
}
