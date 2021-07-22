package testing

import "github.com/ds-test-framework/scheduler/types"

type Action interface {
	Step(*EventWrapper, *MessagePool) []*types.Message
}

type AllowAllAction struct {
}

func (a *AllowAllAction) Step(e *EventWrapper, mPool *MessagePool) []*types.Message {
	msgid, ok := e.Event.MessageID()
	if ok {
		message, ok := mPool.Pick(msgid)
		if ok {
			return []*types.Message{message}
		}
	}
	return []*types.Message{}
}
