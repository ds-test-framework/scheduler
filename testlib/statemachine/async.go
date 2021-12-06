package statemachine

import (
	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/testlib"
	"github.com/ds-test-framework/scheduler/testlib/handlers"
	"github.com/ds-test-framework/scheduler/types"
)

func NewAsyncStateMachineHandler(stateMachine *StateMachine) handlers.HandlerFunc {
	return func(e *types.Event, c *testlib.Context) ([]*types.Message, bool) {
		c.Logger().With(log.LogParams{
			"event_id":   e.ID,
			"event_type": e.TypeS,
		}).Debug("Async state machine handler step")
		stateMachine.step(e, c)
		newState := stateMachine.CurState()
		if newState.Is(FailStateLabel) {
			c.Abort()
		}

		return []*types.Message{}, false
	}
}
