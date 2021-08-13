package types

import (
	"fmt"
	"strings"
)

type MessageSendEventType struct {
	Message *Message
}

func NewMessageSendEventType(message *Message) *MessageSendEventType {
	return &MessageSendEventType{
		Message: message,
	}
}

func (s *MessageSendEventType) Clone() EventType {
	return &MessageSendEventType{
		Message: s.Message,
	}
}

func (s *MessageSendEventType) Type() string {
	return "MessageSendEventType"
}

func (s *MessageSendEventType) String() string {
	return fmt.Sprintf("MessageSend { %s }", s.Message.ID)
}

type MessageReceiveEventType struct {
	Message *Message
}

func NewMessageReceiveEventType(message *Message) *MessageReceiveEventType {
	return &MessageReceiveEventType{
		Message: message,
	}
}

func (r *MessageReceiveEventType) Clone() EventType {
	return &MessageReceiveEventType{
		Message: r.Message,
	}
}

func (r *MessageReceiveEventType) Type() string {
	return "MessageReceiveEventType"
}

func (r *MessageReceiveEventType) String() string {
	return fmt.Sprintf("MessageReceive { %s }", r.Message.ID)
}

type TimeoutStartEventType struct {
	Timeout *ReplicaTimeout
}

func NewTimeoutStartEventType(timeout *ReplicaTimeout) *TimeoutStartEventType {
	return &TimeoutStartEventType{
		Timeout: timeout,
	}
}

func (ts *TimeoutStartEventType) Clone() EventType {
	return &TimeoutStartEventType{
		Timeout: ts.Timeout,
	}
}

func (ts *TimeoutStartEventType) Type() string {
	return "TimeoutStartEventType"
}

func (ts *TimeoutStartEventType) String() string {
	return fmt.Sprintf("TimeoutStart { %s }", ts.Timeout.Type)
}

type TimeoutEndEventType struct {
	Timeout *ReplicaTimeout
}

func NewTimeoutEndEventType(timeout *ReplicaTimeout) *TimeoutEndEventType {
	return &TimeoutEndEventType{
		Timeout: timeout,
	}
}

func (te *TimeoutEndEventType) Clone() EventType {
	return &TimeoutEndEventType{
		Timeout: te.Timeout,
	}
}

func (te *TimeoutEndEventType) Type() string {
	return "TimeoutEndEventType"
}

func (te *TimeoutEndEventType) String() string {
	return fmt.Sprintf("TimeoutEnd { %s }", te.Timeout.Type)
}

type GenericEventType struct {
	Params map[string]string `json:"params"`
	T      string            `json:"type"`
}

func NewGenericEventType(params map[string]string, t string) *GenericEventType {
	return &GenericEventType{
		Params: params,
		T:      t,
	}
}

func (g *GenericEventType) Clone() EventType {
	return &GenericEventType{
		Params: g.Params,
		T:      g.T,
	}
}

func (g *GenericEventType) Type() string {
	return "GenericEvent"
}

func (g *GenericEventType) String() string {
	str := g.T + " {"
	paramS := make([]string, len(g.Params))
	for k, v := range g.Params {
		paramS = append(paramS, fmt.Sprintf(" %s = %s", k, v))
	}
	str += strings.Join(paramS, ",")
	str += " }"
	return str
}
