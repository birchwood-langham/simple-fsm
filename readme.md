# Simple State Machine Framework

This library provides a very basic framework for implementing Finite State Machines (FSM). To use this library, you just need to implement the `State` interface to define each of your states, add the appropriate `Transitions` for each state, then create a state machine and run it, passing in the initial state for your machine.

## State

The `State` interface must be implemented for each state your FSM can be in. In the examples directory, an example turnstyle has been implemented using the state machine.

For our turnstyle, there can be only two states, `Locked` and `Unlocked`, we therefore need to implement the `State` interface for each of these.

### State Interface

The `State` interface is very simple:

```go
type State interface {
  Run(chan Event) (State, error)
  AddTransition(TransitionCheck, TransitionNext)
}
```

The `Run` method is executed by the state machine to perform the actions required for the current state. The `Event` channel is how you pass events to the state for it to process.

When the actions for the state have completed, the `Run` method should return the next `State` or an error. If an error is encountered, the state machine will terminate its loop and send the error up the call stack to be handled.

A state can have multiple transitions so that it may transition to different states given the different circumstances. For example, if you have a shopping basket that is in a Payment authorisation state, if the payment is authorised, you should transition to a completed state, if the payment failed, you may want to transition to a payment entry state, etc.

### Event Interface

A state can process any event that implenents the Event interface. The Event interface is again, very simple requiring just an ID to identify the event and a timestamp to know when the event happened.

```go
type Event interface {
  ID() uuid.UUID
  Timestamp() time.Time
}
```

### Transition

A transition consists of two things, a check function and a transition function. The check function is used to check whether a state is ready to transition to another state. The transition function should create the new state you need to transition to.

A transition is therefore simply defined as

```go
type TransitionCheck func() bool
type TransitionNext func(State) (State, error)

type Transition struct {
  Check TransitionCheck
  Next TransitionNext
}
```

Your state should provide the necessary functions to check whether it is ready to transition to a new state, and functions that transition the current state to the new state.

## Putting it together

For our turnstyle example we need to define the states and events that our turnstyle will respond to. To start with, our turnstyle will be in a locked state, if we try to push the turnstyle, it should not turn. To unlock the turnstyle, we should insert a coin. Once inserted, the turnstyle will be in the unlocked state. If we insert another coin into the turnstyle, it should return the coin and do nothing else. If we push the turnstyle it should turn and then return to the Locked state. If we try to do anything else, it should ignore it.

From the requirements we can see we need two events `InsertCoinEvent` and `PushEvent` and we need two states `Locked` and `Unlocked`.

```go
// Turnstyle Events
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
```

```go
// Locked State
type Locked struct {
  hasCoin     bool
  transitions []fsm.Transition
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

  for _, t := range l.transitions {
    if t.Check() {
      return t.Next(l)
    }
  }

  return l, nil
}

func (l *Locked) Next(_ fsm.State) (fsm.State, error) {
  u := Unlocked{}
  u.AddTransition(u.Pushed, u.Next)

  return u, nil
}

func (l *Locked) HasCoin() bool {
  return l.hasCoin
}

func (l *Locked) AddTransition(check fsm.TransitionCheck, next fsm.TransitionNext) {
  l.transitions = append(l.transitions, fsm.Transition{Check: check, Next: next})
}
```

```go
// Unlocked
type Unlocked struct {
  pushed      bool
  transitions []fsm.Transition
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

  for _, t := range u.transitions {
    if t.Check() {
      return t.Next(u)
    }
  }

  return u, nil
}

func (u *Unlocked) Next(_ fsm.State) (fsm.State, error) {
  l := Locked{}
  l.AddTransition(l.HasCoin, l.Next)

  return l, nil
}

func (u *Unlocked) AddTransition(check fsm.TransitionCheck, next fsm.TransitionNext) {
  u.transitions = append(u.transitions, fsm.Transition{Check: check, Next: next})
}

func (u *Unlocked) Pushed() bool {
  return u.pushed
}
```

Once defined our turnstyle application simply needs to create the `State`s, add the transitions and then run our State Machine:

```go
func main() {
  // the events channel where we publish events to our state machine
  eventsChannel := make(chan fsm.Event)

  // Create our state machine passing in the channel where we will publish events into the machine
  sm := fsm.New(eventsChannel)

  // create our initial state
  locked := Locked{}

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
  eventsChannel <- InsertCoin()
  eventsChannel <- InsertCoin()
  eventsChannel <- Push()
  eventsChannel <- Push()

  close(eventsChannel)
}
```

Running our state machine, we should get the following results:

```shell
2020/06/06 11:50:48 Please insert coin first
2020/06/06 11:50:48 Coin inserted
2020/06/06 11:50:48 Coin already inserted, returning coin
2020/06/06 11:50:48 Turning turnstyle
2020/06/06 11:50:48 Please insert coin first
```
