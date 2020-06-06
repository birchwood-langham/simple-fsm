package fsm

type TransitionCheck func() bool
type TransitionNext func(State) (State, error)

// Transition defines a transition from the given state to the Next state
type Transition struct {
	// Check is the function which is evaluated to determine
	// whether a transition should be performed
	Check TransitionCheck
	// Next is the State to transition to
	Next TransitionNext
}
