package types

type Updater interface {
	Start(layer int)
	Stop()
}
