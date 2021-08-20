package testlib

import (
	"sync"

	"github.com/ds-test-framework/scheduler/types"
)

type Condition func(*Context) bool

type StateAction func(*Context) []*types.Message

type State struct {
	Label       string
	Action      StateAction
	Transitions map[string]Condition
}

type Vars struct {
	vars map[string]interface{}
	lock *sync.Mutex
}

func NewVarSet() *Vars {
	return &Vars{
		vars: make(map[string]interface{}),
		lock: new(sync.Mutex),
	}
}

func (v *Vars) Get(label string) (interface{}, bool) {
	v.lock.Lock()
	defer v.lock.Unlock()

	val, ok := v.vars[label]
	return val, ok
}

func (v *Vars) Set(label string, value interface{}) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.vars[label] = value
}

type StateMachineBuilder struct {
	testCase *TestCase
	curState *State
}

func (s StateMachineBuilder) On(cond Condition, stateLabel string) StateBuilder {
	next, ok := s.testCase.states[stateLabel]
	if !ok {
		next = &State{
			Label:       stateLabel,
			Action:      AllowAllAction,
			Transitions: make(map[string]Condition),
		}
	}
	s.curState.Transitions[next.Label] = cond
	return StateBuilder{
		testCase: s.testCase,
		state:    next,
	}
}

type StateBuilder struct {
	testCase *TestCase
	state    *State
}

func (s StateBuilder) Do(action StateAction) StateMachineBuilder {
	s.state.Action = action
	return StateMachineBuilder{
		testCase: s.testCase,
		curState: s.state,
	}
}
