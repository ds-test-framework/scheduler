package testing

import (
	"sync"
	"time"

	"github.com/ds-test-framework/scheduler/types"
)

type MessagePool struct {
	messages map[string]*types.Message
	lock     *sync.Mutex
}

func NewMessagePool() *MessagePool {
	return &MessagePool{
		messages: make(map[string]*types.Message),
		lock:     new(sync.Mutex),
	}
}

func (p *MessagePool) Pick(id string) (*types.Message, bool) {
	p.lock.Lock()
	defer p.lock.Unlock()

	m, ok := p.messages[id]
	if ok {
		delete(p.messages, id)
	}
	return m, ok
}

func (p *MessagePool) Add(m *types.Message) bool {
	p.lock.Lock()
	defer p.lock.Unlock()
	_, ok := p.messages[m.ID]
	if ok {
		return false
	}
	p.messages[m.ID] = m
	return true
}

type TestCase struct {
	ctx      *types.Context
	Timeout  time.Duration
	states   map[string]*State
	rCtx     *TestCaseCtx
	curState *State
	Name     string
}

func NewTestCase(name string, timeout time.Duration) *TestCase {
	t := &TestCase{
		Name:     name,
		Timeout:  timeout,
		states:   make(map[string]*State),
		curState: nil,
		ctx:      nil,
		rCtx:     NewTestCaseCtx(timeout),
	}
	t.states[startLabel] = NewState(startLabel, &AllowAllAction{})
	t.states[successLabel] = NewState(successLabel, &AllowAllAction{})
	t.states[failLabel] = NewState(failLabel, &AllowAllAction{})
	return t
}

const (
	startLabel   = "start"
	successLabel = "success"
	failLabel    = "fail"
)

func (t *TestCase) StartState() *State {
	return t.states[startLabel]
}

func (t *TestCase) CreateState(label string, action Action) *State {
	state := NewState(label, action)
	t.states[startLabel] = state
	return state
}

func (t *TestCase) WithContext(ctx *types.Context) *TestCase {
	t.ctx = ctx
	return t
}

type TestCaseCtx struct {
	Done    chan bool
	Timeout time.Duration
	once    *sync.Once
}

func NewTestCaseCtx(timeout time.Duration) *TestCaseCtx {
	return &TestCaseCtx{
		Done:    make(chan bool, 1),
		Timeout: timeout,
		once:    new(sync.Once),
	}
}

func (t *TestCaseCtx) SetDone() {
	t.once.Do(func() {
		t.Done <- true
		close(t.Done)
	})
}

func (t *TestCase) Run() *TestCaseCtx {
	return t.rCtx
}

func (t *TestCase) Step(e *types.Event, mPool *MessagePool) []*types.Message {
	messages, transitioned := t.curState.Step(t.ctx, e, mPool)
	if transitioned {
		t.curState = t.curState.Next
		if t.curState.Label == successLabel || t.curState.Label == failLabel {
			t.rCtx.SetDone()
		}
	}
	return messages
}

func (t *TestCase) Assert() bool {
	return t.curState.Label == successLabel
}
