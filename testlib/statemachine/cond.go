package statemachine

import "github.com/ds-test-framework/scheduler/types"

// Condition type for specifying a transition condition
// Condition function is called for every new event that the testing library receives
type Condition func(*types.Event, *Context) bool

func (c Condition) And(other Condition) Condition {
	return func(e *types.Event, ctx *Context) bool {
		if c(e, ctx) && other(e, ctx) {
			return true
		}
		return false
	}
}

func (c Condition) Or(other Condition) Condition {
	return func(e *types.Event, ctx *Context) bool {
		if c(e, ctx) || other(e, ctx) {
			return true
		}
		return false
	}
}

func (c Condition) Not() Condition {
	return func(e *types.Event, ctx *Context) bool {
		return !c(e, ctx)
	}
}

func IsMessageSend() Condition {
	return func(e *types.Event, ctx *Context) bool {
		return e.IsMessageSend()
	}
}

func IsMessageReceive() Condition {
	return func(e *types.Event, ctx *Context) bool {
		return e.IsMessageReceive()
	}
}

func BaseCondition(e *types.Event, c *Context) bool {

	return false
}

func MessageInSet(label string) Condition {
	return func(e *types.Event, c *Context) bool {
		if !e.IsMessageSend() && !e.IsMessageReceive() {
			return false
		}
		mID, _ := e.MessageID()
		set, ok := c.Vars.GetMessageSet(label)
		if !ok {
			return false
		}
		return set.Exists(mID)
	}
}
