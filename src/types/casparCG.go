package types

type InterviewReturn struct{}

type CasparCGClient interface {
	Connect() error
	GetTemplates() ([]string, error)
	PushCGData(template string, data map[string]any) error
}
