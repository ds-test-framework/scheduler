package types

import (
	"fmt"
	"strings"
	"sync"
)

type EventType interface {
	Type() string
	Clone() EventType
	String() string
}

const (
	SendMessageType    = "SendMessage"
	ReceiveMessageType = "ReceiveMessage"
	ReplicaEventType   = "ReplicaEvent"
)

type SendMessage struct {
	msg *Message
}

func NewSendMessageEvent(msg *Message) *SendMessage {
	return &SendMessage{
		msg: msg,
	}
}
func (s *SendMessage) Message() *Message {
	return s.msg
}
func (s *SendMessage) Type() string {
	return SendMessageType
}
func (s *SendMessage) Clone() EventType {
	return &SendMessage{
		msg: s.msg.Clone(),
	}
}
func (s *SendMessage) String() string {
	return "SendMessage(" + s.msg.ID + ")"
}

type ReceiveMessage struct {
	msg *Message
}

func NewReceiveMessageEvent(msg *Message) *ReceiveMessage {
	return &ReceiveMessage{
		msg: msg,
	}
}
func (r *ReceiveMessage) Message() *Message {
	return r.msg
}
func (r *ReceiveMessage) Type() string {
	return ReceiveMessageType
}
func (r *ReceiveMessage) Clone() EventType {
	return &ReceiveMessage{
		msg: r.msg.Clone(),
	}
}
func (r *ReceiveMessage) String() string {
	return "ReceiveMessage(" + r.msg.ID + ")"
}

type ReplicaEvent struct {
	Params map[string]string `json:"params"`
	T      string            `json:"type"`
}

func NewReplicaEvent(t string, params map[string]string) *ReplicaEvent {
	return &ReplicaEvent{
		Params: params,
		T:      t,
	}
}
func (r *ReplicaEvent) Type() string {
	return ReplicaEventType + "(" + r.T + ")"
}
func (r *ReplicaEvent) Clone() EventType {
	return &ReplicaEvent{
		Params: r.Params,
		T:      r.T,
	}
}
func (r *ReplicaEvent) String() string {
	str := r.T + " {"
	paramS := make([]string, len(r.Params))
	for k, v := range r.Params {
		paramS = append(paramS, fmt.Sprintf(" %s = %s", k, v))
	}
	str += strings.Join(paramS, ",")
	str += " }"
	return str
}

// Event encapsulates a message send/receive all necessary information
type Event struct {
	ID        uint        `json:"id"`
	TypeS     string      `json:"type"`
	Type      EventType   `json:"-"`
	Timestamp int64       `json:"timestamp"`
	Replica   ReplicaID   `json:"replica"`
	Prev      uint        `json:"prev"`
	Next      uint        `json:"next"`
	lock      *sync.Mutex `json:"-"`
}

func NewEvent(id uint, replica ReplicaID, t EventType, ts int64) *Event {
	return &Event{
		ID:        id,
		Replica:   replica,
		TypeS:     t.String(),
		Type:      t,
		Timestamp: ts,
		Prev:      0,
		Next:      0,
		lock:      new(sync.Mutex),
	}
}

func (e *Event) UpdatePrev(p *Event) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.Prev = p.ID
}

func (e *Event) UpdateNext(n *Event) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.Next = n.ID
}

func (e *Event) GetNext() uint {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.Next
}

func (e *Event) Clone() *Event {
	e.lock.Lock()
	defer e.lock.Unlock()
	return &Event{
		ID:        e.ID,
		Replica:   e.Replica,
		Type:      e.Type.Clone(),
		Timestamp: e.Timestamp,
		Prev:      e.Prev,
		Next:      e.Next,
		lock:      new(sync.Mutex),
	}
}

func (e *Event) Eq(o *Event) bool {
	return e.ID == o.ID
}

func (e *Event) MessageID() (string, bool) {
	if e.Type.Type() != ReceiveMessageType && e.Type.Type() != SendMessageType {
		return "", false
	}
	var msg *Message
	switch e.Type.Type() {
	case ReceiveMessageType:
		eT := e.Type.(*ReceiveMessage)
		msg = eT.Message()
	case SendMessageType:
		eT := e.Type.(*SendMessage)
		msg = eT.Message()
	}
	return msg.ID, true
}
