package handlers

import (
	"github.com/ds-test-framework/scheduler/testlib"
	"github.com/ds-test-framework/scheduler/types"
)

// HandlerCascade implements Handler
// Executes handlers in the specified order until the event is handled
// If no handler handles the event then the default handler is called
type HandlerCascade struct {
	Handlers       []HandlerFunc
	DefaultHandler HandlerFunc
}

var _ testlib.Handler = &HandlerCascade{}

func defaultHandler(e *types.Event, c *testlib.Context) ([]*types.Message, bool) {
	if !e.IsMessageSend() {
		return []*types.Message{}, true
	}
	mID, _ := e.MessageID()
	message, ok := c.MessagePool.Get(mID)
	if ok {
		return []*types.Message{message}, true
	}
	return []*types.Message{}, true
}

// HandlerCascadeOption changes the parameters of the HandlerCascade
type HandlerCascadeOption func(*HandlerCascade)

// WithDefault changes the HandlerCascade default handler
func WithDefault(d HandlerFunc) HandlerCascadeOption {
	return func(hc *HandlerCascade) {
		hc.DefaultHandler = d
	}
}

// WithStateMachine adds the state machine handler at the head
func WithStateMachine(sm *StateMachine) HandlerCascadeOption {
	return func(hc *HandlerCascade) {
		handlers := []HandlerFunc{NewStateMachineHandler(sm)}
		handlers = append(handlers, hc.Handlers...)
		hc.Handlers = handlers
	}
}

// NewHandlerCascade creates a new cascade handler with the specified options
func NewHandlerCascade(opts ...HandlerCascadeOption) *HandlerCascade {
	h := &HandlerCascade{
		Handlers:       make([]HandlerFunc, 0),
		DefaultHandler: defaultHandler,
	}
	for _, o := range opts {
		o(h)
	}
	return h
}

// AddHandler adds a handler to the cascade
func (c *HandlerCascade) AddHandler(h HandlerFunc) {
	c.Handlers = append(c.Handlers, h)
}

// Then adds a handler to the cascade and returns the cascade
func (c *HandlerCascade) Then(next HandlerFunc) *HandlerCascade {
	c.Handlers = append(c.Handlers, next)
	return c
}

// HandleEvent implements Handler
func (c *HandlerCascade) HandleEvent(e *types.Event, ctx *testlib.Context) []*types.Message {
	for _, h := range c.Handlers {
		ret, ok := h(e, ctx)
		if ok {
			return ret
		}
	}
	ret, _ := c.DefaultHandler(e, ctx)
	return ret
}

func (c *HandlerCascade) Name() string {
	return "HandlerCascade"
}
