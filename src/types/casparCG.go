package types

type InterviewReturn struct{}

type CasparCGClient interface {
	Connect() error
	GetTemplates() ([]string, error)

	// Control functions for CG templates
	PushCGData(template string, layer, channel int, data map[string]any, posX, posY *int, sizeX, sizeY *float64) error
	StopCGData(template string, layer, channel int) error
}
