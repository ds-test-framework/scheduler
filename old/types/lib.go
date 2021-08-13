package types

import "sync"

type ReplicaID string

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

func (m *MessageWrapper) Clone() Clonable {
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

func (r *Replica) Clone() Clonable {
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

type Clonable interface {
	Clone() Clonable
}
