package types

type InterviewReturn struct{}

type CasparCGClient interface {
	Connect() error
}
