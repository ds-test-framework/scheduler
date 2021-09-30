package testlib

import (
	"sync"
	"time"

	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
)

type Handler interface {
	HandleEvent(*Context) []*types.Message
	Name() string
}

type HandlerFunc func(*Context) []*types.Message

type GenericHandler struct {
	handlerFunc HandlerFunc
}

func NewGenericHandler(f HandlerFunc) *GenericHandler {
	return &GenericHandler{
		handlerFunc: f,
	}
}

func (g *GenericHandler) HandleEvent(c *Context) []*types.Message {
	return g.handlerFunc(c)
}

func (g *GenericHandler) Name() string {
	return "GenericHandler"
}

type DoNothingHandler struct {
}

func (d *DoNothingHandler) HandleEvent(_ *Context) []*types.Message {
	return []*types.Message{}
}

func (d *DoNothingHandler) Name() string {
	return "DoNothing"
}

// Type assertion
var _ Handler = &DoNothingHandler{}
var _ Handler = &GenericHandler{}

// TestCase represents a unit test case
type TestCase struct {
	// Name name of the testcase
	Name string
	// Timeout maximum duration of the testcase execution
	Timeout time.Duration
	// Setup function called prior to initiation of the execution
	Setup   func(*Context) error
	Handler Handler
	success bool
	// Logger to log information
	Logger *log.Logger

	doneCh chan string
	once   *sync.Once
}

func defaultSetupFunc(c *Context) error {
	return nil
}

// NewTestCase instantiates a TestCase based on the parameters specified
// The new testcase has three states by default.
// - Start state where the execution starts from
// - Fail state that can be used to fail the testcase
// - Success state that can be used to indicate a success of the testcase
func NewTestCase(name string, timeout time.Duration, handler Handler) *TestCase {
	return &TestCase{
		Name:    name,
		Timeout: timeout,
		Handler: handler,
		Setup:   defaultSetupFunc,
		success: false,
		doneCh:  make(chan string, 1),
		once:    new(sync.Once),
	}
}

func (t *TestCase) Succeed() {
	t.success = true
}

func (t *TestCase) End() {
	t.once.Do(func() {
		close(t.doneCh)
	})
}

func (t *TestCase) Abort() {
	t.success = false
	t.once.Do(func() {
		close(t.doneCh)
	})
}

// Step is called to execute a step of the testcase with a new event
func (t *TestCase) step(c *Context) []*types.Message {
	return t.Handler.HandleEvent(c)
}

// Assert returns true if the testcase statemachine is in the success state
// or if the curState is marked as success
func (t *TestCase) Assert() bool {
	return t.success
}

// SetupFunc can be used to set the setup function
func (t *TestCase) SetupFunc(setupFunc func(*Context) error) {
	t.Setup = setupFunc
}
