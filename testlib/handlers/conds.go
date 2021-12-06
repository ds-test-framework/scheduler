package handlers

import (
	"github.com/ds-test-framework/scheduler/testlib"
	"github.com/ds-test-framework/scheduler/types"
)

type Condition func(e *types.Event, c *testlib.Context) bool

func (c Condition) And(other Condition) Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		if c(e, ctx) && other(e, ctx) {
			return true
		}
		return false
	}
}

func (c Condition) Or(other Condition) Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		if c(e, ctx) || other(e, ctx) {
			return true
		}
		return false
	}
}

func (c Condition) Not() Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		return !c(e, ctx)
	}
}

func IsMessageSend() Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		return e.IsMessageSend()
	}
}

func IsMessageReceive() Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		return e.IsMessageReceive()
	}
}
