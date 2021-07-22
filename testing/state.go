package testing

import (
	"sync"

	"github.com/ds-test-framework/scheduler/types"
)

// EventWrapper is to be passed to the condition checker
// Encapsulates the new event that has been added and the entire graph
// To facilitate condition check both as a stream of new events
// and as a function of the current state of the graph
type EventWrapper struct {
	Event *types.Event
	Graph *types.EventGraph
}

// Condition Abstracts a predicate over the cur graph state
type Condition interface {
	// Should return true if the predicate is satisfied, false otherwise
	Check(*EventWrapper) bool
}

// Transition indicates an edge out of a state
// A labelled edge which is to be taken if the condition is satisfied
type Transition struct {
	To   *State
	Cond Condition
}

// State of a state machine with a map of possible transitions
type State struct {
	Label       string
	Transitions map[string]Transition
	Action      Action
	Next        *State
}

// NewState creates a state with the specified label and transitions
func NewState(label string, action Action) *State {
	return &State{
		Label:       label,
		Action:      action,
		Transitions: make(map[string]Transition),
		Next:        nil,
	}
}

// HandleEvent checks if the event allows for a transition from the given state
// Returns true if the current new event allows the state machine to transition, false otherwise
func (s *State) handleEvent(ctx *types.Context, e *types.Event) bool {
	var wg sync.WaitGroup
	results := make(chan *State, len(s.Transitions))

	eventW := &EventWrapper{
		Event: e,
		Graph: ctx.EventGraph.GetCurGraph(),
	}
	for _, t := range s.Transitions {
		wg.Add(1)
		go func(transition Transition) {
			if transition.Cond.Check(eventW) {
				results <- transition.To
			} else {
				results <- nil
			}
			wg.Done()
		}(t)
	}
	wg.Wait()
	for i := 0; i < len(s.Transitions); i++ {
		r := <-results
		if r != nil && s.Next == nil {
			s.Next = r
		}
	}
	return s.Next != nil
}

func (s *State) Step(ctx *types.Context, e *types.Event, mPool *MessagePool) ([]*types.Message, bool) {
	return s.Action.Step(&EventWrapper{
		Event: e,
		Graph: ctx.EventGraph.GetCurGraph(),
	}, mPool), s.handleEvent(ctx, e)
}

func (s *State) Upon(cond Condition, next *State) *State {
	s.Transitions[next.Label] = Transition{
		Cond: cond,
		To:   next,
	}
	return next
}

type VarSet struct {
	vars map[string]interface{}
	lock *sync.Mutex
}

func NewVarSet() *VarSet {
	return &VarSet{
		vars: make(map[string]interface{}),
		lock: new(sync.Mutex),
	}
}

func (v *VarSet) Get(label string) (interface{}, bool) {
	v.lock.Lock()
	defer v.lock.Unlock()

	val, ok := v.vars[label]
	return val, ok
}

func (v *VarSet) Set(label string, value interface{}) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.vars[label] = value
}

// TODO:
// - Need a state machine structure to store states and global variables of the machine.
// - Each state should be associated with an action, that decides to update global variables, or deliver messages from a pool
// - Need a state machine builder. Most natural pattern to create test cases
