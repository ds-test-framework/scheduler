package handlers

import (
	"encoding/json"
	"sync"

	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/testlib"
	"github.com/ds-test-framework/scheduler/types"
)

const (
	StartStateLabel   = "startState"
	FailStateLabel    = "failState"
	SuccessStateLabel = "successState"
)

type StateMachineBuilder struct {
	stateMachine *StateMachine
	curState     *State
}

// On can be used to create a transition relation between states based on the specified condition
func (s StateMachineBuilder) On(cond Condition, stateLabel string) StateMachineBuilder {
	next, ok := s.stateMachine.getState(stateLabel)
	if !ok {
		next = s.stateMachine.newState(stateLabel)
	}
	s.curState.Transitions[next.Label] = cond
	return StateMachineBuilder{
		stateMachine: s.stateMachine,
		curState:     next,
	}
}

func (s StateMachineBuilder) MarkSuccess() StateMachineBuilder {
	s.curState.Success = true
	return s
}

// State of the testcase state machine
type State struct {
	Label       string               `json:"label"`
	Transitions map[string]Condition `json:"-"`
	Success     bool                 `json:"success"`
}

func (s *State) Is(l string) bool {
	return s.Label == l
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
	for to := range s.Transitions {
		transitions[i] = to
		i++
	}
	keyvals["transitions"] = transitions
	return json.Marshal(keyvals)
}

type run struct {
	curState    *State
	lock        *sync.Mutex
	transitions []string
}

func newRun(start *State) *run {
	return &run{
		curState:    start,
		lock:        new(sync.Mutex),
		transitions: []string{start.Label},
	}
}

func (r *run) Transition(to *State) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.curState = to
	r.transitions = append(r.transitions, to.Label)
}

func (r *run) GetTransitions() []string {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.transitions
}

func (r *run) CurState() *State {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.curState
}

type StateMachine struct {
	states map[string]*State
	run    *run
}

func NewStateMachine() *StateMachine {
	m := &StateMachine{
		states: make(map[string]*State),
	}
	startState := &State{
		Label:       StartStateLabel,
		Success:     false,
		Transitions: make(map[string]Condition),
	}
	m.states[StartStateLabel] = startState
	m.run = newRun(startState)
	m.states[FailStateLabel] = &State{
		Label:       FailStateLabel,
		Success:     false,
		Transitions: make(map[string]Condition),
	}
	m.states[SuccessStateLabel] = &State{
		Label:       SuccessStateLabel,
		Success:     true,
		Transitions: make(map[string]Condition),
	}
	return m
}

func (s *StateMachine) Builder() StateMachineBuilder {
	return StateMachineBuilder{
		stateMachine: s,
		curState:     s.states[StartStateLabel],
	}
}

func (s *StateMachine) CurState() *State {
	return s.run.CurState()
}

func (s *StateMachine) Transition(to string) {
	state, ok := s.getState(to)
	if ok {
		s.run.Transition(state)
	}
}

func (s *StateMachine) getState(label string) (*State, bool) {
	state, ok := s.states[label]
	return state, ok
}

func (s *StateMachine) newState(label string) *State {
	cur, ok := s.states[label]
	if ok {
		return cur
	}
	newState := &State{
		Label:       label,
		Transitions: make(map[string]Condition),
		Success:     false,
	}
	s.states[label] = newState
	return newState
}

func (s *StateMachine) step(e *types.Event, c *testlib.Context) {
	state := s.run.CurState()
	for to, t := range state.Transitions {
		if t(e, c) {
			next, ok := s.states[to]
			if ok {
				c.Logger().With(log.LogParams{
					"state": to,
				}).Info("Testcase transistioned")
				s.run.Transition(next)
				c.Vars.Set("curState", to)
			}
		}
	}
}

func (s *StateMachine) InSuccessState() bool {
	return s.run.CurState().Success
}

func InState(state string) Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		curState, ok := c.Vars.GetString("curState")
		return ok && curState == state
	}
}

func NewStateMachineHandler(stateMachine *StateMachine) HandlerFunc {
	return func(e *types.Event, c *testlib.Context) ([]*types.Message, bool) {
		c.Logger().With(log.LogParams{
			"event_id":   e.ID,
			"event_type": e.TypeS,
		}).Debug("Async state machine handler step")
		stateMachine.step(e, c)
		newState := stateMachine.CurState()
		if newState.Is(FailStateLabel) {
			c.Abort()
		}

		return []*types.Message{}, false
	}
}
