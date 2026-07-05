package types

type UpdateJob interface {
	Start() error
	Stop()
}
