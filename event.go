package fsm

import "time"

// EventType defines what an event is
type EventType string

// Event is something that happens in the state machine
// State machines will process events and determine whether
// the state should transition to a new state
type Event interface {
	// Timestamp of the event
	Timestamp() time.Time
	// EventType returns the type of event this is
	EventType() EventType
}
