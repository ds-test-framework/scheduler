package types

import "sync"

type EventType = string

const (
	Send    EventType = "Send"
	Receive EventType = "Receive"
)

type Event struct {
	ID        uint
	Type      EventType
	MsgID     string
	Timestamp int64
	Replica   uint
	Prev      uint
	Next      uint
	lock      *sync.Mutex
}

func NewEvent(id, replica uint, t EventType, ts int64, msg string) *Event {
	return &Event{
		ID:        id,
		Replica:   replica,
		Type:      t,
		MsgID:     msg,
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
		Type:      e.Type,
		MsgID:     e.MsgID,
		Timestamp: e.Timestamp,
		Prev:      e.Prev,
		Next:      e.Next,
		lock:      new(sync.Mutex),
	}
}

func (e *Event) Eq(o *Event) bool {
	return e.ID == o.ID
}

type Message struct {
	Type         string
	ID           string
	From         uint
	To           uint
	Weight       int
	Timeout      bool
	SendEvent    uint
	ReceiveEvent uint
	lock         *sync.Mutex
}

func NewMessage(t, id string, from, to uint, w int, timeout bool) *Message {
	return &Message{
		Type:         t,
		ID:           id,
		From:         from,
		To:           to,
		Weight:       w,
		Timeout:      timeout,
		SendEvent:    0,
		ReceiveEvent: 0,
		lock:         new(sync.Mutex),
	}
}

func (m *Message) GetSendEvent() uint {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.SendEvent
}

func (m *Message) GetReceiveEvent() uint {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.ReceiveEvent
}

func (m *Message) UpdateSendEvent(e uint) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.SendEvent = e
}

func (m *Message) UpdateReceiveEvent(e uint) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.ReceiveEvent = e
}

func (m *Message) Clone() *Message {
	m.lock.Lock()
	defer m.lock.Unlock()
	return &Message{
		ID:           m.ID,
		Type:         m.Type,
		From:         m.From,
		To:           m.To,
		Weight:       m.Weight,
		Timeout:      m.Timeout,
		SendEvent:    m.SendEvent,
		ReceiveEvent: m.ReceiveEvent,
		lock:         new(sync.Mutex),
	}
}

type MessageWrapper struct {
	Run int
	Msg *Message
}
