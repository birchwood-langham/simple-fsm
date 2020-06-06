package main

import (
	"log"
	"time"

	"github.com/birchwood-langham/fsm"
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

		switch e.(type) {
		case PushEvent:
			log.Println("Please insert coin first")
			continue
		case InsertCoinEvent:
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

type InsertCoinEvent struct {
	timestamp time.Time
}

func (a InsertCoinEvent) Timestamp() time.Time {
	return a.timestamp
}

func InsertCoin() InsertCoinEvent {
	return InsertCoinEvent{timestamp: time.Now()}
}

type PushEvent struct {
	timestamp time.Time
}

func (p PushEvent) Timestamp() time.Time {
	return p.timestamp
}

func Push() PushEvent {
	return PushEvent{timestamp: time.Now()}
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

		switch e.(type) {
		case InsertCoinEvent:
			log.Println("Coin already inserted, returning coin")
			continue
		case PushEvent:
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

type PullEvent struct {
	timestamp time.Time
}

func (p PullEvent) Timestamp() time.Time {
	return p.timestamp
}

func Pull() PullEvent {
	return PullEvent{timestamp: time.Now()}
}

func main() {
	eventsChannel := make(chan fsm.Event)

	sm := fsm.New(eventsChannel)

	locked := LockedState()
	unlocked := UnlockedState()

	locked.AddTransition(locked.HasCoin, unlocked)
	unlocked.AddTransition(unlocked.Pushed, locked)

	go sm.Run(locked)

	eventsChannel <- Push()
	eventsChannel <- Pull()
	eventsChannel <- InsertCoin()
	eventsChannel <- InsertCoin()
	eventsChannel <- Push()

	time.Sleep(time.Second)

	close(eventsChannel)
}
