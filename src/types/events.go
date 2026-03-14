package types

type WailsPayload struct {
	Identifier string `json:"identifier"`
	Value      any    `json:"value"`
}

type EventIdentifier string

const (
	EventIdentifierCasparCGKeepAlive EventIdentifier = "CasparCGKeepAlive"
)

type CasparCGKeepAlive struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	IsAlive bool   `json:"isAlive"`
}

func (e CasparCGKeepAlive) GetIdentifier() EventIdentifier {
	return EventIdentifierCasparCGKeepAlive
}

func (e CasparCGKeepAlive) GetData() any {
	return e
}

type Event interface {
	GetIdentifier() EventIdentifier
	GetData() any
}

// EventProcessor defines an interface for processing events.
// It provides methods to push new events and to listen for incoming events.
type EventProcessor interface {
	// Push pushes an event to all listeners
	Push(event Event) error
	// Listen subscribes to the channel where all events will be sent to
	Listen() <-chan Event
	// CloseChannel closes a given channel.
	// This function should be called when a listener no longer wants to receive events
	CloseChannel(ch <-chan Event)

	// Cancel closes the Event Processor
	Cancel()
}
