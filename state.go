package fsm

// State represents a state of the machine
type State interface {
	// Run executes the actions for the given state
	// It should process the events coming in from the
	// event channel, and then check the transitions to determine
	// whether a transition to another state should be performed
	Run(chan Event) (State, error)
	// WithTransitions adds transitions for the state and returns
	// the updated state
	WithTransitions(...Transition) State
}
