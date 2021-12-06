package handlers

import (
	"github.com/ds-test-framework/scheduler/testlib"
	"github.com/ds-test-framework/scheduler/types"
)

type IfThenHandler struct {
	cond Condition
	then HandlerFunc
}

func If(cond Condition) *IfThenHandler {
	return &IfThenHandler{
		cond: cond,
	}
}

func (i *IfThenHandler) Then(h HandlerFunc) HandlerFunc {
	return func(e *types.Event, c *testlib.Context) ([]*types.Message, bool) {
		if i.cond(e, c) {
			return i.then(e, c)
		}
		return []*types.Message{}, false
	}
}
