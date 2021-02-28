package fsm

import (
	"time"

	"github.com/oklog/ulid/v2"
)

// Event is something that happens in the state machine
// State machines will process events and determine whether
// the state should transition to a new state
type Event interface {
	// ID is the unique identifier for the event
	ID() ulid.ULID
	// Timestamp of the event
	Timestamp() time.Time
}
