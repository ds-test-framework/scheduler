package testlib

import (
	"sync"
	"time"

	"github.com/ds-test-framework/scheduler/context"
	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
)

var (
	startStateLabel   = "start"
	successStateLabel = "success"
	failureStateLabel = "failure"
)

type Context struct {
	MessagePool *types.MessageStore
	Replicas    *types.ReplicaStore
	CurEvent    *types.Event
	EventDAG    *types.EventDAG
	Vars        *Vars
	CurState    *State

	latestEvents map[types.ReplicaID]*types.Event
	sends        map[string]*types.Event
	lock         *sync.Mutex
}

func NewContext(c *context.RootContext) *Context {
	return &Context{
		MessagePool: c.MessageStore,
		Replicas:    c.Replicas,
		CurEvent:    nil,
		EventDAG:    types.NewEventDag(),
		Vars:        NewVarSet(),
		CurState:    nil,

		latestEvents: make(map[types.ReplicaID]*types.Event),
		sends:        make(map[string]*types.Event),
		lock:         new(sync.Mutex),
	}
}

func (c *Context) setEvent(e *types.Event) {
	c.lock.Lock()
	defer c.lock.Unlock()
	latest, ok := c.latestEvents[e.Replica]
	parents := make([]*types.Event, 0)
	if ok {
		parents = append(parents, latest)
	}
	c.latestEvents[e.Replica] = e

	switch e.Type.(type) {
	case *types.MessageReceiveEventType:
		eventType := e.Type.(*types.MessageReceiveEventType)
		send, ok := c.sends[eventType.MessageID]
		if ok {
			parents = append(parents, send)
		}
	case *types.MessageSendEventType:
		eventType := e.Type.(*types.MessageSendEventType)
		c.sends[eventType.MessageID] = e
	}
	c.EventDAG.AddNode(e, parents)
	c.CurEvent = e
}

type runState struct {
	curState *State
	lock     *sync.Mutex
}

func (r *runState) CurState() *State {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.curState
}

func (r *runState) Transition(s *State) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.curState = s
}

type TestCase struct {
	Name    string
	Timeout time.Duration
	Setup   func(*Context) error
	states  map[string]*State
	Logger  *log.Logger

	run    *runState
	doneCh chan string
	once   *sync.Once
}

func defaultSetupFunc(c *Context) error {
	return nil
}

func AllowAllAction(c *Context) []*types.Message {
	event := c.CurEvent
	switch event.Type.(type) {
	case *types.MessageSendEventType:
		eventType := event.Type.(*types.MessageSendEventType)
		message, ok := c.MessagePool.Get(eventType.MessageID)
		if ok {
			return []*types.Message{message}
		}
	}
	return []*types.Message{}
}

func NewTestCase(name string, timeout time.Duration) *TestCase {
	t := &TestCase{
		Name:    name,
		Timeout: timeout,
		states:  make(map[string]*State),
		Setup:   defaultSetupFunc,
		doneCh:  make(chan string, 1),
		once:    new(sync.Once),
	}
	startState := &State{
		Label:       startStateLabel,
		Action:      AllowAllAction,
		Transitions: make(map[string]Condition),
	}
	t.states[startStateLabel] = startState
	t.run = &runState{
		curState: startState,
		lock:     new(sync.Mutex),
	}
	t.states[successStateLabel] = &State{
		Label:       successStateLabel,
		Action:      AllowAllAction,
		Transitions: make(map[string]Condition),
	}
	t.states[failureStateLabel] = &State{
		Label:       failureStateLabel,
		Action:      AllowAllAction,
		Transitions: make(map[string]Condition),
	}
	return t
}

func (t *TestCase) Builder() StateMachineBuilder {
	return StateMachineBuilder{
		testCase: t,
		curState: t.Start(),
	}
}

func (t *TestCase) Start() *State {
	return t.states[startStateLabel]
}

func (t *TestCase) Success() *State {
	return t.states[successStateLabel]
}

func (t *TestCase) Fail() *State {
	return t.states[failureStateLabel]
}

func (t *TestCase) Step(c *Context) []*types.Message {
	c.CurState = t.run.CurState()
	result := t.run.CurState().Action(c)
	for label, cond := range t.run.CurState().Transitions {
		if cond(c) {
			nextState, ok := t.states[label]
			if ok {
				t.run.Transition(nextState)
			}
		}
	}
	curState := t.run.CurState()
	if curState.Label == successStateLabel || curState.Label == failureStateLabel {
		t.once.Do(func() {
			t.doneCh <- curState.Label
		})
	}
	return result
}

func (t *TestCase) Assert() bool {
	return t.run.CurState().Label == successStateLabel
}

func (t *TestCase) SetupFunc(setupFunc func(*Context) error) {
	t.Setup = setupFunc
}
