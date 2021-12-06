package testlib

import (
	"sync"
	"time"

	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
)

type Handler interface {
	HandleEvent(*types.Event, *Context) []*types.Message
	Name() string
}

type DoNothingHandler struct {
}

func (d *DoNothingHandler) HandleEvent(_ *types.Event, _ *Context) []*types.Message {
	return []*types.Message{}
}

func (d *DoNothingHandler) Name() string {
	return "DoNothing"
}

// Type assertion
var _ Handler = &DoNothingHandler{}

// TestCase represents a unit test case
type TestCase struct {
	// Name name of the testcase
	Name string
	// Timeout maximum duration of the testcase execution
	Timeout time.Duration
	// setup function called prior to initiation of the execution
	setup    func(*Context) error
	assertFn func(*Context) bool
	Handler  Handler
	aborted  bool
	// Logger to log information
	Logger *log.Logger

	doneCh chan string
	once   *sync.Once
}

func defaultSetupFunc(c *Context) error {
	return nil
}

func defaultAssertFunc(c *Context) bool {
	return false
}

// NewTestCase instantiates a TestCase based on the parameters specified
// The new testcase has three states by default.
// - Start state where the execution starts from
// - Fail state that can be used to fail the testcase
// - Success state that can be used to indicate a success of the testcase
func NewTestCase(name string, timeout time.Duration, handler Handler) *TestCase {
	return &TestCase{
		Name:     name,
		Timeout:  timeout,
		Handler:  handler,
		setup:    defaultSetupFunc,
		assertFn: defaultAssertFunc,
		aborted:  false,
		doneCh:   make(chan string, 1),
		once:     new(sync.Once),
	}
}

func (t *TestCase) End() {
	t.once.Do(func() {
		close(t.doneCh)
	})
}

func (t *TestCase) Abort() {
	t.aborted = true
	t.once.Do(func() {
		close(t.doneCh)
	})
}

// Step is called to execute a step of the testcase with a new event
func (t *TestCase) step(e *types.Event, c *Context) []*types.Message {
	return t.Handler.HandleEvent(e, c)
}

// SetupFunc can be used to set the setup function
func (t *TestCase) SetupFunc(setupFunc func(*Context) error) {
	t.setup = setupFunc
}

type AssertFunc func(*Context) bool

func (t *TestCase) AssertFn(fn AssertFunc) {
	t.assertFn = fn
}

func (t *TestCase) assert(c *Context) bool {
	return t.assertFn(c)
}
