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

// Context struct is passed to the calls of StateAction and Condition
// encapsulates all information needed by the StateAction and Condition to function
type Context struct {
	// MessagePool reference to an instance of the MessageStore
	MessagePool *types.MessageStore
	// Replicas reference to the replica store
	Replicas *types.ReplicaStore
	// CurEvent is the event that was processed latest
	CurEvent *types.Event
	// EventDAG is the directed acyclic graph all prior events
	EventDAG *types.EventDAG
	// Vars is a generic key value store to facilate maintaining auxilliary information
	// during the execution of a testcase
	Vars *Vars
	// CurState is the current state of the testcase state machine
	CurState *State

	testcase     *TestCase
	latestEvents map[types.ReplicaID]*types.Event
	sends        map[string]*types.Event
	lock         *sync.Mutex
}

// NewContext instantiates a Context from the RootContext
func NewContext(c *context.RootContext, testcase *TestCase) *Context {
	return &Context{
		MessagePool: c.MessageStore,
		Replicas:    c.Replicas,
		CurEvent:    nil,
		EventDAG:    types.NewEventDag(),
		Vars:        NewVarSet(),
		CurState:    nil,

		testcase:     testcase,
		latestEvents: make(map[types.ReplicaID]*types.Event),
		sends:        make(map[string]*types.Event),
		lock:         new(sync.Mutex),
	}
}

// Transition can be used by actions to force a transition from the current state
func (c *Context) Transition(newstate string) {
	s, ok := c.testcase.states[newstate]
	if ok {
		c.testcase.run.Transition(s)
	}
}

// Logger returns the logger for the current testcase
func (c *Context) Logger() *log.Logger {
	return c.testcase.Logger
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

// TestCase represents a unit test case
type TestCase struct {
	// Name name of the testcase
	Name string
	// Timeout maximum duration of the testcase execution
	Timeout time.Duration
	// Setup function called prior to initiation of the execution
	Setup  func(*Context) error
	states map[string]*State
	// Logger to log information
	Logger *log.Logger

	run    *runState
	doneCh chan string
	once   *sync.Once
}

func defaultSetupFunc(c *Context) error {
	return nil
}

// Default Action of the state
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

// NewTestCase instantiates a TestCase based on the parameters specified
// The new testcase has three states by default.
// - Start state where the execution starts from
// - Fail state that can be used to fail the testcase
// - Success state that can be used to indicate a success of the testcase
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

// Builder returns a StateMachineBuilder which can be used for construction the testcase statemachine
func (t *TestCase) Builder() StateMachineBuilder {
	return StateMachineBuilder{
		testCase: t,
		curState: t.Start(),
	}
}

// Start returns the Start state of the testcase
func (t *TestCase) Start() *State {
	return t.states[startStateLabel]
}

// Success returns the success state
func (t *TestCase) Success() *State {
	return t.states[successStateLabel]
}

// Fail returns the fail state
func (t *TestCase) Fail() *State {
	return t.states[failureStateLabel]
}

// Step is called to execute a step of the testcase with a new event
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

// Assert returns true if the testcase statemachine is in the success state
func (t *TestCase) Assert() bool {
	return t.run.CurState().Label == successStateLabel
}

// SetupFunc can be used to set the setup function
func (t *TestCase) SetupFunc(setupFunc func(*Context) error) {
	t.Setup = setupFunc
}

func (t *TestCase) States() map[string]*State {
	return t.states
}
