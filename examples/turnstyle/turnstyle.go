package main

import (
	"log"
	"time"

	"github.com/birchwood-langham/fsm"
	"github.com/google/uuid"
)

type Locked struct {
	hasCoin     bool
	transitions []fsm.Transition
}

func LockedState() *Locked {
	l := &Locked{}

	return l
}

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

func (l *Locked) checkTransitions() (fsm.State, error) {
	for _, t := range l.transitions {
		if t.Check() {
			return t.Next(l)
		}
	}

	return l, nil
}

func (l *Locked) Next(_ fsm.State) (fsm.State, error) {
	u := UnlockedState()
	u.AddTransition(u.Pushed, u.Next)

	return u, nil
}

func (l *Locked) HasCoin() bool {
	return l.hasCoin
}

func (l *Locked) AddTransition(check fsm.TransitionCheck, next fsm.TransitionNext) {
	l.transitions = append(l.transitions, fsm.Transition{Check: check, Next: next})
}

type InsertCoinEvent struct {
	id        uuid.UUID
	timestamp time.Time
}

func (a InsertCoinEvent) Timestamp() time.Time {
	return a.timestamp
}

func (a InsertCoinEvent) ID() uuid.UUID {
	return a.id
}

func InsertCoin() InsertCoinEvent {
	return InsertCoinEvent{
		id:        uuid.New(),
		timestamp: time.Now(),
	}
}

type PushEvent struct {
	id        uuid.UUID
	timestamp time.Time
}

func (p PushEvent) ID() uuid.UUID {
	return p.id
}

func (p PushEvent) Timestamp() time.Time {
	return p.timestamp
}

func Push() PushEvent {
	return PushEvent{
		id:        uuid.New(),
		timestamp: time.Now(),
	}
}

type Unlocked struct {
	pushed      bool
	transitions []fsm.Transition
}

func UnlockedState() *Unlocked {
	u := &Unlocked{}
	return u
}

func (u *Unlocked) Pushed() bool {
	return u.pushed
}

func (u *Unlocked) AddTransition(check fsm.TransitionCheck, next fsm.TransitionNext) {
	u.transitions = append(u.transitions, fsm.Transition{Check: check, Next: next})
}

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

func (u *Unlocked) checkTransitions() (fsm.State, error) {
	for _, t := range u.transitions {
		if t.Check() {
			return t.Next(u)
		}
	}

	return u, nil
}

func (u *Unlocked) Next(_ fsm.State) (fsm.State, error) {
	l := LockedState()
	l.AddTransition(l.HasCoin, l.Next)

	return l, nil
}

type PullEvent struct {
	id        uuid.UUID
	timestamp time.Time
}

func (p PullEvent) Timestamp() time.Time {
	return p.timestamp
}

func (p PullEvent) ID() uuid.UUID {
	return p.id
}

func Pull() PullEvent {
	return PullEvent{
		id:        uuid.New(),
		timestamp: time.Now(),
	}
}

func main() {
	eventsChannel := make(chan fsm.Event)

	sm := fsm.New(eventsChannel)

	locked := LockedState()

	locked.AddTransition(locked.HasCoin, locked.Next)

	statusCh := sm.Run(locked)

	defer func() {
		status := <-statusCh

		if status.Error != nil {
			log.Fatalf("state machine terminated in error: %v", status.Error)
		}
	}()

	eventsChannel <- Push()
	eventsChannel <- Pull()
	eventsChannel <- InsertCoin()
	eventsChannel <- InsertCoin()
	eventsChannel <- Push()
	eventsChannel <- Push()

	close(eventsChannel)
}
