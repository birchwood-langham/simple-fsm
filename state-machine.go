package fsm

type StateMachine struct {
	eventsChannel chan Event
}

func New(events chan Event) *StateMachine {
	stateMachine := StateMachine{
		eventsChannel: events,
	}

	return &stateMachine
}

func (sm StateMachine) Run(initialState State) (err error) {
	for s := initialState; s != nil; {
		if s, err = s.Run(sm.eventsChannel); err != nil {
			return
		}
	}

	return
}
