package types

type SheetsData interface {
	Start()
	Close()

	GetCountdown() *SimpleCountdown
	GetCountdownDuration() *SimpleCountdown
	GetDJCountdown() *DJCountdown
	GetDJCountdownDuration() *DJCountdown

	GetLowerThird01() *LowerThird
	GetLowerThird02() *LowerThird
	GetLowerThirdDJ() *LowerThird

	GetCurrentSchedule() []*ScheduleRow
}
