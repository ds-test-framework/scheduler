package statemachine

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

// State of the testcase state machine
type State struct {
	Label       string               `json:"label"`
	Transitions map[string]Condition `json:"-"`
	Success     bool                 `json:"success"`

	handler EventHandler
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
		handler:     nil,
	}
	s.states[label] = newState
	return newState
}

func (s *StateMachine) step(e *types.Event, c *Context) {
	state := s.run.CurState()
	for to, t := range state.Transitions {
		if t(e, c) {
			next, ok := s.states[to]
			if ok {
				c.Logger().With(log.LogParams{
					"state": to,
				}).Info("Testcase transistioned")
				s.run.Transition(next)
			}
		}
	}
}

type EventHandler func(*types.Event, *Context) ([]*types.Message, bool)

func defaultSendHandler(e *types.Event, c *Context) ([]*types.Message, bool) {
	if !e.IsMessageSend() {
		return []*types.Message{}, true
	}
	messageID, _ := e.MessageID()
	message, ok := c.MessagePool.Get(messageID)
	if ok {
		return []*types.Message{message}, true
	}
	return []*types.Message{}, true
}

type Context struct {
	*testlib.Context
	StateMachine *StateMachine
}

func wrapContext(c *testlib.Context, m *StateMachine) *Context {
	return &Context{
		Context:      c,
		StateMachine: m,
	}
}
