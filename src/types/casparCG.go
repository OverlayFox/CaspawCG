package types

type Sizing struct {
	PosX  int     `json:"posX"`
	PosY  int     `json:"posY"`
	SizeX float64 `json:"sizeX"`
	SizeY float64 `json:"sizeY"`
}

type InterviewReturn struct{}

type CasparCGClient interface {
	Connect() error
	GetTemplates() ([]string, error)

	// Control functions for CG templates
	PushCGData(template string, layer, channel int, data map[string]any, sizing Sizing) error
	StopCGData(template string, layer, channel int) error
}
