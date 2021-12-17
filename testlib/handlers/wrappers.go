package handlers

import (
	"fmt"

	"github.com/ds-test-framework/scheduler/testlib"
	"github.com/ds-test-framework/scheduler/types"
)

type CountWrapper struct {
	counterFunc func(*types.Event, *testlib.Context) (*testlib.Counter, bool)
}

func Count(label string) *CountWrapper {
	return &CountWrapper{
		counterFunc: func(_ *types.Event, c *testlib.Context) (*testlib.Counter, bool) {
			return c.Vars.GetCounter(label)
		},
	}
}

func CountTo(label string) *CountWrapper {
	return &CountWrapper{
		counterFunc: func(e *types.Event, c *testlib.Context) (*testlib.Counter, bool) {
			message, ok := c.GetMessage(e)
			if !ok {
				return nil, false
			}
			counter, ok := c.Vars.GetCounter(fmt.Sprintf("%s_%s", label, message.To))
			if !ok {
				return nil, false
			}
			return counter, true
		},
	}
}

type SetWrapper struct {
	setFunc func(*types.Event, *testlib.Context) (*types.MessageStore, bool)
}

func Set(label string) *SetWrapper {
	return &SetWrapper{
		setFunc: func(e *types.Event, c *testlib.Context) (*types.MessageStore, bool) {
			return c.Vars.GetMessageSet(label)
		},
	}
}

func (s *SetWrapper) Count() *CountWrapper {
	return &CountWrapper{
		counterFunc: func(e *types.Event, c *testlib.Context) (*testlib.Counter, bool) {
			set, ok := s.setFunc(e, c)
			if !ok {
				return nil, false
			}
			counter := testlib.NewCounter()
			counter.SetValue(set.Size())
			return counter, true
		},
	}
}
