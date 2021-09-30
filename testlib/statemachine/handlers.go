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

func doNothingHandler(c *Context) ([]*types.Message, bool) {
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
	return func(c *Context) ([]*types.Message, bool) {
		if h.cond(c) {
			return h.then(c)
		} else if h.el != nil {
			return h.el(c)
		}
		return []*types.Message{}, false
	}
}
