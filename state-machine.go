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

func (sm StateMachine) Run(initialState State) chan Status {
	statusChannel := make(chan Status, 1)
	go sm.run(initialState, statusChannel)
	return statusChannel
}

func (sm StateMachine) run(initialState State, statusChannel chan<- Status) {
	var err error

	for s := initialState; s != nil; {
		if s, err = s.Run(sm.eventsChannel); err != nil {
			statusChannel <- Status{Error: err}
			return
		}
	}

	statusChannel <- Status{Error: nil}
}
