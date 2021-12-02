package statemachine

import "github.com/ds-test-framework/scheduler/types"

type IfThenElseHandler struct {
	cond Condition
	then EventHandler
	el   EventHandler
}

func If(c Condition) *IfThenElseHandler {
	return &IfThenElseHandler{
		cond: c,
		then: doNothingHandler,
		el:   nil,
	}
}

func doNothingHandler(e *types.Event, c *Context) ([]*types.Message, bool) {
	return []*types.Message{}, false
}

func (h *IfThenElseHandler) Then(handler EventHandler) EventHandler {
	h.then = handler
	return h.Handler()
}

func (h *IfThenElseHandler) ThenElse(then EventHandler, el EventHandler) EventHandler {
	h.then = then
	h.el = el
	return h.Handler()
}

func (h *IfThenElseHandler) Handler() EventHandler {
	return func(e *types.Event, c *Context) ([]*types.Message, bool) {
		if h.cond(e, c) {
			return h.then(e, c)
		} else if h.el != nil {
			return h.el(e, c)
		}
		return []*types.Message{}, false
	}
}

func DeliverMessage(e *types.Event, c *Context) ([]*types.Message, bool) {
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

func DontDeliverMessage(e *types.Event, c *Context) ([]*types.Message, bool) {
	if !e.IsMessageSend() {
		return []*types.Message{}, false
	}
	messageID, _ := e.MessageID()
	_, ok := c.MessagePool.Get(messageID)
	return []*types.Message{}, ok
}

func RecordMessage(label string) EventHandler {
	return func(e *types.Event, c *Context) ([]*types.Message, bool) {
		if !e.IsMessageSend() {
			return []*types.Message{}, false
		}
		messageID, _ := e.MessageID()
		message, ok := c.MessagePool.Get(messageID)
		if ok {
			c.Vars.Set(label, message)
			return []*types.Message{}, true
		}
		return []*types.Message{}, false
	}
}

func AddToSet(label string) EventHandler {
	return func(e *types.Event, c *Context) ([]*types.Message, bool) {
		if !e.IsMessageSend() {
			return []*types.Message{}, false
		}
		messageID, _ := e.MessageID()
		message, ok := c.MessagePool.Get(messageID)
		if ok {
			set, ok := c.Vars.GetMessageSet(label)
			if !ok {
				c.Vars.NewMessageSet(label)
				set, _ = c.Vars.GetMessageSet(label)
			}
			set.Add(message)
			return []*types.Message{}, true
		}
		return []*types.Message{}, false
	}
}
