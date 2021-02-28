package main

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"

	fsm "github.com/birchwood-langham/simple-fsm"
)

var once sync.Once
var entropy *ulid.MonotonicEntropy

func NewULID() ulid.ULID {
	once.Do(func() {
		entropy = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	})

	for {
		id, err := ulid.New(ulid.Timestamp(time.Now()), entropy)
		if err != nil {
			continue
		}
		return id
	}
}

// InsertCoinEvent represents an insert coin action
type InsertCoinEvent struct {
	id        ulid.ULID
	timestamp time.Time
}

// Timestamp is the time when the event was created
func (a InsertCoinEvent) Timestamp() time.Time {
	return a.timestamp
}

// ID is the unique identifier for the event
func (a InsertCoinEvent) ID() ulid.ULID {
	return a.id
}

// InsertCoin constructs an InsertCoinEvent
func InsertCoin() InsertCoinEvent {
	return InsertCoinEvent{
		id:        NewULID(),
		timestamp: time.Now(),
	}
}

// PushEvent represents a push action on the turnstyle
type PushEvent struct {
	id        ulid.ULID
	timestamp time.Time
}

// ID is the unique identifier for the event
func (p PushEvent) ID() ulid.ULID {
	return p.id
}

// Timestamp is the time when the event was created
func (p PushEvent) Timestamp() time.Time {
	return p.timestamp
}

// Push constructs a PushEvent
func Push() PushEvent {
	return PushEvent{
		id:        NewULID(),
		timestamp: time.Now(),
	}
}

// PullEvent represents a pull action on the turnstyle
type PullEvent struct {
	id        ulid.ULID
	timestamp time.Time
}

// Timestamp is the time the event was created
func (p PullEvent) Timestamp() time.Time {
	return p.timestamp
}

// ID returns the ulid for the PullEvent
func (p PullEvent) ID() ulid.ULID {
	return p.id
}

// Pull creates a PullEvent with a new ulid and the current timestamp
func Pull() PullEvent {
	return PullEvent{
		id:        NewULID(),
		timestamp: time.Now(),
	}
}

// Locked models the locked state of a turnstyle
type Locked struct {
	hasCoin     bool
	transitions []fsm.Transition
}

// LockedState constructs a locked state object
func LockedState() *Locked {
	l := &Locked{}

	return l
}

// Run executes the actions required for the locked state
func (l *Locked) Run(incoming chan fsm.Event) (fsm.State, error) {
	e, ok := <-incoming

	if !ok {
		return nil, nil
	}

	switch e.(type) {
	case PushEvent:
		log.Println("Please insert coin first")
	case InsertCoinEvent:
		log.Println("Coin inserted")

		l.hasCoin = true
	default:
		log.Println("Unknown event, ignoring")
	}

	return l.checkTransitions()
}

// checkTransitions checks the transitions for the locked state
func (l *Locked) checkTransitions() (fsm.State, error) {
	for _, t := range l.transitions {
		if t.Check() {
			return t.Next(l)
		}
	}

	return l, nil
}

// Next is the transition function from Locked to Unlocked state
func (l *Locked) Next(_ fsm.State) (fsm.State, error) {
	u := UnlockedState()
	u.AddTransition(u.Pushed, u.Next)

	return u, nil
}

// HasCoin is the transition check function for the Locked state
func (l *Locked) HasCoin() bool {
	return l.hasCoin
}

// AddTransition adds a transition to the Locked state
func (l *Locked) AddTransition(check fsm.TransitionCheck, next fsm.TransitionNext) {
	l.transitions = append(l.transitions, fsm.Transition{Check: check, Next: next})
}

// Unlocked represents the unlocked state of a turnstyle
type Unlocked struct {
	pushed      bool
	transitions []fsm.Transition
}

// UnlockedState contructor
func UnlockedState() *Unlocked {
	u := &Unlocked{}
	return u
}

// Pushed is the transition check function for the Unlocked state
func (u *Unlocked) Pushed() bool {
	return u.pushed
}

// AddTransition appends a transition to the list of transtions for the state
func (u *Unlocked) AddTransition(check fsm.TransitionCheck, next fsm.TransitionNext) {
	u.transitions = append(u.transitions, fsm.Transition{Check: check, Next: next})
}

// Run executes the actions for the Unlocked state
func (u *Unlocked) Run(incoming chan fsm.Event) (fsm.State, error) {
	e, ok := <-incoming

	if !ok {
		return nil, nil
	}

	switch e.(type) {
	case InsertCoinEvent:
		log.Println("Coin already inserted, returning coin")
	case PushEvent:
		log.Println("Turning turnstyle")

		u.pushed = true
	default:
		log.Println("Unknown event, ignoring")
	}

	return u.checkTransitions()
}

// checkTransition checks each of the transtions for the state to determine if
// the state is ready to transition to the next state
func (u *Unlocked) checkTransitions() (fsm.State, error) {
	for _, t := range u.transitions {
		if t.Check() {
			return t.Next(u)
		}
	}

	return u, nil
}

// Next is the transtion function that changes the Unlocked state to the Locked state
func (u *Unlocked) Next(_ fsm.State) (fsm.State, error) {
	l := LockedState()
	l.AddTransition(l.HasCoin, l.Next)

	return l, nil
}

func main() {
	// the events channel where we publish events to our state machine
	eventsChannel := make(chan fsm.Event)

	// Create our state machine passing in the channel where we will publish events into the machine
	sm := fsm.New(eventsChannel)

	// create our initial state
	locked := LockedState()

	// add the transitions we expect the states to handle, our state definition provides our TransactionCheck function HasCoin
	// and the Next() function defines how we transition to the next state
	locked.AddTransition(locked.HasCoin, locked.Next)

	// Calling the Run() method starts our state machine. The state machine is run in a separate thread, and a channel is returned
	// so that we can listen for the Status of our state machine when it completes.
	statusCh := sm.Run(locked)

	// our state machine is running in a separate thread, we want to make sure that it the main thread
	// doesn't terminate before the other thread can finish processing
	defer func() {
		status := <-statusCh

		if status.Error != nil {
			log.Fatalf("state machine terminated in error: %v", status.Error)
		}
	}()

	// We send the events to our state machine, as the events are processed, the machine will switch
	// between the two states
	eventsChannel <- Push()
	eventsChannel <- Pull()
	eventsChannel <- InsertCoin()
	eventsChannel <- InsertCoin()
	eventsChannel <- Push()
	eventsChannel <- Push()

	// close the events channel
	close(eventsChannel)
}
