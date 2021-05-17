package types

import "sync"

type EventType string
type ReplicaID string

const (
	Send    EventType = "Send"
	Receive EventType = "Receive"
)

// Event encapsulates a message send/receive all necessary information
type Event struct {
	ID        uint
	Type      EventType
	MsgID     string
	Timestamp int64
	Replica   ReplicaID
	Prev      uint
	Next      uint
	lock      *sync.Mutex
}

func NewEvent(id uint, replica ReplicaID, t EventType, ts int64, msg string) *Event {
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

// Message encapsulates communication between the nodes/replicas
type Message struct {
	Type         string      `json:"type"`
	ID           string      `json:"id"`
	From         ReplicaID   `json:"from"`
	To           ReplicaID   `json:"to"`
	Weight       int         `json:"-"`
	Timeout      bool        `json:"-"`
	SendEvent    uint        `json:"-"`
	ReceiveEvent uint        `json:"-"`
	Msg          []byte      `json:"msg"`
	Intercept    bool        `json:"intercept"`
	lock         *sync.Mutex `json:"-"`
}

func NewMessage(t, id string, from, to ReplicaID, w int, timeout bool, msg []byte, intercept bool) *Message {
	return &Message{
		Type:         t,
		ID:           id,
		From:         from,
		To:           to,
		Weight:       w,
		Timeout:      timeout,
		Msg:          msg,
		SendEvent:    0,
		ReceiveEvent: 0,
		Intercept:    intercept,
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
	if m == nil {
		return nil
	}
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
		Intercept:    m.Intercept,
		lock:         new(sync.Mutex),
		Msg:          m.Msg,
	}
}

// MessageWrapper wraps around message annotating it with the run number
type MessageWrapper struct {
	Run int
	Msg *Message
}

func (m *MessageWrapper) Clone() *MessageWrapper {
	if m == nil {
		return nil
	}
	return &MessageWrapper{
		Run: m.Run,
		Msg: m.Msg.Clone(),
	}
}

type Replica struct {
	ID    ReplicaID              `json:"id"`
	Addr  string                 `json:"addr"`
	Info  map[string]interface{} `json:"info,omitempty"`
	Ready bool                   `json:"ready"`
}

func (r *Replica) Clone() *Replica {
	if r == nil {
		return nil
	}
	return &Replica{
		ID:    r.ID,
		Addr:  r.Addr,
		Info:  r.Info,
		Ready: r.Ready,
	}
}

type Timeout struct {
	Type     string    `json:"type"`
	Duration int       `json:"duration"`
	Replica  ReplicaID `json:"replica"`
}
