package statemachine

import (
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
	ctx := wrapContext(c, a.StateMachine)
	result := make([]*types.Message, 0)
	handled := false
	for _, handler := range a.EventHandlers {
		messages, h := handler(ctx)
		if h {
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
	if newState.Success {
		c.Success()
	} else if newState.Label == FailStateLabel {
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
