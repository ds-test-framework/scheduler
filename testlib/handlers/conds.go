package handlers

import (
	"github.com/ds-test-framework/scheduler/testlib"
	"github.com/ds-test-framework/scheduler/types"
)

// Condition type to define predicates over the current event or the history of events
type Condition func(e *types.Event, c *testlib.Context) bool

// And to create boolean conditional expressions
func (c Condition) And(other Condition) Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		if c(e, ctx) && other(e, ctx) {
			return true
		}
		return false
	}
}

// Or to create boolean conditional expressions
func (c Condition) Or(other Condition) Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		if c(e, ctx) || other(e, ctx) {
			return true
		}
		return false
	}
}

// Not to create boolean conditional expressions
func (c Condition) Not() Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		return !c(e, ctx)
	}
}

// IsMessageSend condition returns true if the event is a message send event
func IsMessageSend() Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		return e.IsMessageSend()
	}
}

// IsMessageReceive condition returns true if the event is a message receive event
func IsMessageReceive() Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		return e.IsMessageReceive()
	}
}

func IsEventType(t string) Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		eType, ok := e.Type.(*types.GenericEventType)
		if !ok {
			return false
		}
		return eType.T == t
	}
}

func IsMessageType(t string) Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		message, ok := c.GetMessage(e)
		if !ok {
			return false
		}
		return message.Type == t
	}
}

func IsMessageTo(to types.ReplicaID) Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		message, ok := c.GetMessage(e)
		if !ok {
			return false
		}
		return message.To == to
	}
}

func IsMessageFrom(from types.ReplicaID) Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		message, ok := c.GetMessage(e)
		if !ok {
			return false
		}
		return message.From == from
	}
}

func (c *CountWrapper) Lt(val func(*types.Event, *testlib.Context) (int, bool)) Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		counter, ok := c.counterFunc(e, ctx)
		if !ok {
			return false
		}
		v, ok := val(e, ctx)
		if !ok {
			return false
		}
		return counter.Value() < v
	}
}

func (c *CountWrapper) Gt(val func(*types.Event, *testlib.Context) (int, bool)) Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		counter, ok := c.counterFunc(e, ctx)
		if !ok {
			return false
		}
		v, ok := val(e, ctx)
		if !ok {
			return false
		}
		return counter.Value() > v
	}
}

func (c *CountWrapper) Eq(val func(*types.Event, *testlib.Context) (int, bool)) Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		counter, ok := c.counterFunc(e, ctx)
		if !ok {
			return false
		}
		v, ok := val(e, ctx)
		if !ok {
			return false
		}
		return counter.Value() == v
	}
}

func (c *CountWrapper) Leq(val func(*types.Event, *testlib.Context) (int, bool)) Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		counter, ok := c.counterFunc(e, ctx)
		if !ok {
			return false
		}
		v, ok := val(e, ctx)
		if !ok {
			return false
		}
		return counter.Value() <= v
	}
}

func (c *CountWrapper) Geq(val func(*types.Event, *testlib.Context) (int, bool)) Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		counter, ok := c.counterFunc(e, ctx)
		if !ok {
			return false
		}
		v, ok := val(e, ctx)
		if !ok {
			return false
		}
		return counter.Value() >= v
	}
}

func (c *CountWrapper) LtValue(val int) Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		counter, ok := c.counterFunc(e, ctx)
		if !ok {
			return false
		}
		return counter.Value() < val
	}
}

func (c *CountWrapper) GtValue(val int) Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		counter, ok := c.counterFunc(e, ctx)
		if !ok {
			return false
		}
		return counter.Value() > val
	}
}

func (c *CountWrapper) EqValue(val int) Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		counter, ok := c.counterFunc(e, ctx)
		if !ok {
			return false
		}
		return counter.Value() == val
	}
}

func (c *CountWrapper) LeqValue(val int) Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		counter, ok := c.counterFunc(e, ctx)
		if !ok {
			return false
		}
		return counter.Value() <= val
	}
}

func (c *CountWrapper) GeqValue(val int) Condition {
	return func(e *types.Event, ctx *testlib.Context) bool {
		counter, ok := c.counterFunc(e, ctx)
		if !ok {
			return false
		}
		return counter.Value() >= val
	}
}

func (s *SetWrapper) Contains() Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		set, ok := s.setFunc(e, c)
		if !ok {
			return false
		}
		message, ok := c.GetMessage(e)
		if !ok {
			return false
		}
		return set.Exists(message.ID)
	}
}
