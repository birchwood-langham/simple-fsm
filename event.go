package fsm

import (
	"time"

	"github.com/google/uuid"
)

// Event is something that happens in the state machine
// State machines will process events and determine whether
// the state should transition to a new state
type Event interface {
	// ID is the unique identifier for the event
	ID() uuid.UUID
	// Timestamp of the event
	Timestamp() time.Time
}
