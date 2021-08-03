package testing

import (
	"fmt"

	"github.com/ds-test-framework/scheduler/types"
)

type Action interface {
	Step(*EventWrapper, *MessagePool, *VarSet) []*types.Message
}

type AllowAllAction struct {
}

func (a *AllowAllAction) Step(e *EventWrapper, mPool *MessagePool, _ *VarSet) []*types.Message {
	msgid, ok := e.Event.MessageID()
	fmt.Printf("Handling event %s", e.Event.TypeS)
	if ok {
		fmt.Printf("A message event with id: %s", msgid)
		message, ok := mPool.Pick(msgid)
		if ok {
			return []*types.Message{message}
		}
	}
	return []*types.Message{}
}
