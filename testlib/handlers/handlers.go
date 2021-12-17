package handlers

import (
	"github.com/ds-test-framework/scheduler/testlib"
	"github.com/ds-test-framework/scheduler/types"
)

type IfThenHandler struct {
	cond     Condition
	handlers []HandlerFunc
}

func If(cond Condition) *IfThenHandler {
	return &IfThenHandler{
		cond:     cond,
		handlers: make([]HandlerFunc, 0),
	}
}

func (i *IfThenHandler) Then(handler HandlerFunc, rest ...HandlerFunc) HandlerFunc {
	i.handlers = append(i.handlers, handler)
	i.handlers = append(i.handlers, rest...)
	return func(e *types.Event, c *testlib.Context) ([]*types.Message, bool) {
		if i.cond(e, c) {
			result := make([]*types.Message, 0)
			for _, h := range i.handlers {
				messages, ok := h(e, c)
				if ok {
					result = append(result, messages...)
				}
			}
			return result, true
		}
		return []*types.Message{}, false
	}
}

// DeliverMessage returns the message if the event is a message send event
func DeliverMessage() HandlerFunc {
	return func(e *types.Event, c *testlib.Context) ([]*types.Message, bool) {
		if !e.IsMessageSend() {
			return []*types.Message{}, false
		}
		messageID, _ := e.MessageID()
		message, ok := c.MessagePool.Get(messageID)
		if ok {
			return []*types.Message{message}, true
		}
		return []*types.Message{}, false
	}
}

// DontDeliverMessage returns ([], true) if the event is a message send event
func DontDeliverMessage() HandlerFunc {
	return func(e *types.Event, c *testlib.Context) ([]*types.Message, bool) {
		if !e.IsMessageSend() {
			return []*types.Message{}, false
		}
		messageID, _ := e.MessageID()
		_, ok := c.MessagePool.Get(messageID)
		return []*types.Message{}, ok
	}
}

func (c *CountWrapper) Incr() HandlerFunc {
	return func(e *types.Event, ctx *testlib.Context) ([]*types.Message, bool) {
		counter, ok := c.counterFunc(e, ctx)
		if !ok {
			return []*types.Message{}, false
		}
		counter.Incr()
		return []*types.Message{}, false
	}
}

func (s *SetWrapper) Store() HandlerFunc {
	return func(e *types.Event, c *testlib.Context) ([]*types.Message, bool) {
		set, ok := s.setFunc(e, c)
		if !ok {
			return []*types.Message{}, false
		}
		message, ok := c.GetMessage(e)
		if !ok {
			return []*types.Message{}, false
		}
		set.Add(message)
		return []*types.Message{}, false
	}
}

func (s *SetWrapper) DeliverAll() HandlerFunc {
	return func(e *types.Event, c *testlib.Context) ([]*types.Message, bool) {
		set, ok := s.setFunc(e, c)
		if !ok {
			return []*types.Message{}, false
		}
		result := set.Iter()
		set.RemoveAll()
		return result, true
	}
}

func RecordMessageAs(label string) HandlerFunc {
	return func(e *types.Event, c *testlib.Context) ([]*types.Message, bool) {
		message, ok := c.GetMessage(e)
		if !ok {
			return []*types.Message{}, false
		}
		c.Vars.Set(label, message)
		return []*types.Message{}, false
	}
}
