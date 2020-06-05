package main

import (
	"log"
	"time"

	"github.com/birchwood-langham/fsm"
)

const (
	InsertCoin fsm.EventType = "Insert Coin"
	Push       fsm.EventType = "Push Turnstyle"
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
	for {
		e, ok := <-incoming

		if !ok {
			return nil, nil
		}

		switch e.EventType() {
		case Push:
			log.Println("Please insert coin first")
			continue
		case InsertCoin:
			log.Println("Coin inserted")

			l.hasCoin = true
		default:
			log.Println("Unknown event, ignoring")
		}

		for _, t := range l.transitions {
			if t.Check() {
				return t.Next, nil
			}
		}
	}
}

func (l *Locked) HasCoin() bool {
	return l.hasCoin
}

func (l *Locked) AddTransition(check func() bool, next fsm.State) {
	l.transitions = append(l.transitions, fsm.Transition{Check: check, Next: next})
}

type AddMoney struct{}

func (AddMoney) EventType() fsm.EventType {
	return InsertCoin
}

type PushTurnstyle struct{}

func (PushTurnstyle) EventType() fsm.EventType {
	return Push
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

func (u *Unlocked) AddTransition(check func() bool, next fsm.State) {
	u.transitions = append(u.transitions, fsm.Transition{Check: check, Next: next})
}

func (u *Unlocked) Run(incoming chan fsm.Event) (fsm.State, error) {
	for {
		e, ok := <-incoming

		if !ok {
			return nil, nil
		}

		switch e.EventType() {
		case InsertCoin:
			log.Println("Coin already inserted, returning coin")
			continue
		case Push:
			log.Println("Turning turnstyle")

			u.pushed = true
		default:
			log.Println("Unknown event, ignoring")
			continue
		}

		for _, t := range u.transitions {
			if t.Check() {
				return t.Next, nil
			}
		}
	}
}

func main() {
	eventsChannel := make(chan fsm.Event)

	sm := fsm.New(eventsChannel)

	locked := LockedState()
	unlocked := UnlockedState()

	locked.AddTransition(locked.HasCoin, unlocked)
	unlocked.AddTransition(unlocked.Pushed, locked)

	go sm.Run(locked)

	eventsChannel <- PushTurnstyle{}
	eventsChannel <- AddMoney{}
	eventsChannel <- AddMoney{}
	eventsChannel <- PushTurnstyle{}

	time.Sleep(time.Second)

	close(eventsChannel)
}
