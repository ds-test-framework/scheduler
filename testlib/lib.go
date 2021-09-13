package testlib

import (
	"encoding/json"
	"sync"

	"github.com/ds-test-framework/scheduler/types"
)

// Condition type for specifying a transition condition
// Condition function is called for every new event that the testing library receives
type Condition func(*Context) bool

// StateAction encapsulates the function of a state
// Called for every new event
type StateAction func(*Context) []*types.Message

// State of the testcase state machine
type State struct {
	Label       string               `json:"label"`
	Action      StateAction          `json:"-"`
	Transitions map[string]Condition `json:"-"`
}

// Eq returns true if the two state labels are the same
func (s *State) Eq(other *State) bool {
	return s.Label == other.Label
}

func (s *State) MarshalJSON() ([]byte, error) {
	keyvals := make(map[string]interface{})
	keyvals["label"] = s.Label
	transitions := make([]string, len(s.Transitions))
	i := 0
	for to, _ := range s.Transitions {
		transitions[i] = to
		i++
	}
	keyvals["transitions"] = transitions
	return json.Marshal(keyvals)
}

// Vars is a dictionary for storing auxilliary state during the execution of the testcase
// Vars is stored in the context passed to actions and conditions
type Vars struct {
	vars map[string]interface{}
	lock *sync.Mutex
}

// NewVarSet instantiates Vars
func NewVarSet() *Vars {
	return &Vars{
		vars: make(map[string]interface{}),
		lock: new(sync.Mutex),
	}
}

// Get returns the value stored  of the specified label
// the second return argument is false if the label does not exist
func (v *Vars) Get(label string) (interface{}, bool) {
	v.lock.Lock()
	defer v.lock.Unlock()

	val, ok := v.vars[label]
	return val, ok
}

// Set the value at the specified label
func (v *Vars) Set(label string, value interface{}) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.vars[label] = value
}

// Exists returns true if there is a variable of the specified key
func (v *Vars) Exists(label string) bool {
	v.lock.Lock()
	defer v.lock.Unlock()
	_, ok := v.vars[label]
	return ok
}

// StateMachineBuilder implements a builder pattern for creating the testcase state machine
type StateMachineBuilder struct {
	testCase *TestCase
	curState *State
}

// On can be used to create a transition relation between states based on the specified condition
func (s StateMachineBuilder) On(cond Condition, stateLabel string) StateBuilder {
	next, ok := s.testCase.states[stateLabel]
	if !ok {
		next = &State{
			Label:       stateLabel,
			Action:      AllowAllAction,
			Transitions: make(map[string]Condition),
		}
		s.testCase.states[next.Label] = next
	}
	s.curState.Transitions[next.Label] = cond
	return StateBuilder{
		testCase: s.testCase,
		state:    next,
	}
}

// StateBuilder implements a builder pattern for specifying the action in the state machine
type StateBuilder struct {
	testCase *TestCase
	state    *State
}

// Do should be called to initialize the action of the current state
func (s StateBuilder) Action(action StateAction) StateMachineBuilder {
	s.state.Action = action
	return StateMachineBuilder{
		testCase: s.testCase,
		curState: s.state,
	}
}
