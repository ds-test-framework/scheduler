package statemachine

import (
	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/testlib"
	"github.com/ds-test-framework/scheduler/types"
)

var _ testlib.Handler = &AsyncStateMachineHandler{}

type AsyncStateMachineHandler struct {
	EventHandlers []EventHandler
	StateMachine  *StateMachine
}

func NewAsyncStateMachineHandler(stateMachine *StateMachine) *AsyncStateMachineHandler {
	return &AsyncStateMachineHandler{
		StateMachine:  stateMachine,
		EventHandlers: make([]EventHandler, 0),
	}
}

func (a *AsyncStateMachineHandler) HandleEvent(c *testlib.Context) []*types.Message {
	c.Logger().With(log.LogParams{
		"event_id":   c.CurEvent.ID,
		"event_type": c.CurEvent.TypeS,
	}).Debug("Async state machine handler step")
	ctx := wrapContext(c, a.StateMachine)
	result := make([]*types.Message, 0)
	handled := false
	for i, handler := range a.EventHandlers {
		messages, h := handler(ctx)
		if h {
			c.Logger().With(log.LogParams{
				"handler_index": i,
			}).Debug("Event handled by handler")
			handled = true
			result = messages
			break
		}
	}
	if !handled {
		result, _ = defaultSendHandler(ctx)
	}

	a.StateMachine.step(ctx)
	newState := a.StateMachine.CurState()
	if newState.Is(FailStateLabel) {
		c.Abort()
	}

	return result
}

func (a *AsyncStateMachineHandler) Name() string {
	return "AsyncStateMachineHandler"
}

func (a *AsyncStateMachineHandler) Finalize() {
}

func (a *AsyncStateMachineHandler) AddEventHandler(handler EventHandler) {
	a.EventHandlers = append(a.EventHandlers, handler)
}
