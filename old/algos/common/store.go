package common

import (
	"errors"
	"sync"

	"github.com/ds-test-framework/scheduler/types"
)

// MsgStore
type MsgStore struct {
	counters       map[types.ReplicaID]int
	messages       map[string]*types.Message
	toBeDispatched map[types.ReplicaID][]*types.Message

	updateCh chan types.ReplicaID
	lock     *sync.Mutex
}

func NewMsgStore(updateCh chan types.ReplicaID) *MsgStore {
	return &MsgStore{
		counters:       make(map[types.ReplicaID]int),
		messages:       make(map[string]*types.Message),
		toBeDispatched: make(map[types.ReplicaID][]*types.Message),
		updateCh:       updateCh,
		lock:           new(sync.Mutex),
	}
}

func (m *MsgStore) Add(msg *types.Message) {
	m.lock.Lock()
	m.messages[msg.ID] = msg
	m.lock.Unlock()
}

func (m *MsgStore) Mark(msgID string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	msg, ok := m.messages[msgID]
	if ok {
		_, ok = m.counters[msg.To]
		if !ok {
			m.counters[msg.To] = 0
		}
		m.counters[msg.To] += 1
		_, ok = m.toBeDispatched[msg.To]
		if !ok {
			m.toBeDispatched[msg.To] = make([]*types.Message, 0)
		}
		m.toBeDispatched[msg.To] = append(m.toBeDispatched[msg.To], msg)
		go func() {
			m.updateCh <- msg.To
		}()
	}
}

func (m *MsgStore) Count(p types.ReplicaID) int {
	m.lock.Lock()
	defer m.lock.Unlock()

	count, ok := m.counters[p]
	if !ok {
		return 0
	}
	return count
}

func (m *MsgStore) Fetch(p types.ReplicaID, limit int) ([]*types.Message, int) {
	m.lock.Lock()
	defer m.lock.Unlock()

	count, ok := m.counters[p]
	if !ok {
		return make([]*types.Message, 0), 0
	}

	var remaining int
	result := m.toBeDispatched[p][:limit]
	m.toBeDispatched[p] = m.toBeDispatched[p][limit:]
	if limit > count {
		m.counters[p] = 0
		remaining = 0
	} else {
		m.counters[p] = count - limit
		remaining = count - limit
	}

	return result, remaining
}

func (m *MsgStore) FetchOne(p types.ReplicaID) (*types.Message, error) {
	res, _ := m.Fetch(p, 1)
	if len(res) == 0 {
		return nil, errors.New("no messages for peer")
	}
	return res[0], nil
}

func (m *MsgStore) Exists(msgID string) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	_, ok := m.messages[msgID]
	return ok
}

func (m *MsgStore) Reset() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.counters = make(map[types.ReplicaID]int)
	m.toBeDispatched = make(map[types.ReplicaID][]*types.Message)
	m.messages = make(map[string]*types.Message)
}
