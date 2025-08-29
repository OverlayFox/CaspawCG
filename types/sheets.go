package types

type SheetsHandler interface {
	Start()
	Close()

	GetData(layer int, templatePath TemplatePaths) ([]byte, error)
	GetCurrentSchedule(startIndex int) ([]byte, error)
}
