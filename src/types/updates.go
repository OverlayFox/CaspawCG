package types

type UpdateJob interface {
	Start() error
	TriggerNow() error
	Stop()
}
