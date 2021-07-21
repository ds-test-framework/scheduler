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
	SendMessageTypeS    = "SendMessage"
	ReceiveMessageTypeS = "ReceiveMessage"
	ReplicaEventTypeS   = "ReplicaEvent"
)

type SendMessageEventType struct {
	msg *Message
}

func NewSendMessageEventType(msg *Message) *SendMessageEventType {
	return &SendMessageEventType{
		msg: msg,
	}
}
func (s *SendMessageEventType) Message() *Message {
	return s.msg
}
func (s *SendMessageEventType) Type() string {
	return SendMessageTypeS
}
func (s *SendMessageEventType) Clone() EventType {
	return &SendMessageEventType{
		msg: s.msg.Clone(),
	}
}
func (s *SendMessageEventType) String() string {
	return "SendMessage(" + s.msg.ID + ")"
}

type ReceiveMessageEventType struct {
	msg *Message
}

func NewReceiveMessageEventType(msg *Message) *ReceiveMessageEventType {
	return &ReceiveMessageEventType{
		msg: msg,
	}
}
func (r *ReceiveMessageEventType) Message() *Message {
	return r.msg
}
func (r *ReceiveMessageEventType) Type() string {
	return ReceiveMessageTypeS
}
func (r *ReceiveMessageEventType) Clone() EventType {
	return &ReceiveMessageEventType{
		msg: r.msg.Clone(),
	}
}
func (r *ReceiveMessageEventType) String() string {
	return "ReceiveMessage(" + r.msg.ID + ")"
}

type ReplicaEventType struct {
	Params map[string]string `json:"params"`
	T      string            `json:"type"`
}

func NewReplicaEventType(t string, params map[string]string) *ReplicaEventType {
	return &ReplicaEventType{
		Params: params,
		T:      t,
	}
}
func (r *ReplicaEventType) Type() string {
	return ReplicaEventTypeS + "(" + r.T + ")"
}
func (r *ReplicaEventType) Clone() EventType {
	return &ReplicaEventType{
		Params: r.Params,
		T:      r.T,
	}
}
func (r *ReplicaEventType) String() string {
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

func (e *Event) Clone() Clonable {
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
	if e.Type.Type() != ReceiveMessageTypeS && e.Type.Type() != SendMessageTypeS {
		return "", false
	}
	var msg *Message
	switch e.Type.Type() {
	case ReceiveMessageTypeS:
		eT := e.Type.(*ReceiveMessageEventType)
		msg = eT.Message()
	case SendMessageTypeS:
		eT := e.Type.(*SendMessageEventType)
		msg = eT.Message()
	}
	return msg.ID, true
}
