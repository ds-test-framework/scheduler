package handlers

import (
	"github.com/ds-test-framework/scheduler/testlib"
	"github.com/ds-test-framework/scheduler/types"
)

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

type HandlerCascadeOption func(*HandlerCascade)

func WithDefault(d HandlerFunc) HandlerCascadeOption {
	return func(hc *HandlerCascade) {
		hc.DefaultHandler = d
	}
}

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

func (h HandlerFunc) Then(next HandlerFunc) *HandlerCascade {
	c := NewHandlerCascade()
	c.AddHandler(h)
	c.AddHandler(next)
	return c
}

func (c *HandlerCascade) AddHandler(h HandlerFunc) {
	c.Handlers = append(c.Handlers, h)
}

func (c *HandlerCascade) Then(next HandlerFunc) *HandlerCascade {
	c.Handlers = append(c.Handlers, next)
	return c
}

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
