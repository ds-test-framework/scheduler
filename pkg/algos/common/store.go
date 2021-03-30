package common

import (
	"errors"
	"sync"
)

type msgStore struct {
	counters       map[PeerID]int
	messages       map[string]*InterceptedMessage
	toBeDispatched map[PeerID][]*InterceptedMessage

	updateCh chan PeerID
	lock     *sync.Mutex
}

func newMsgStore(updateCh chan PeerID) *msgStore {
	return &msgStore{
		counters:       make(map[PeerID]int),
		messages:       make(map[string]*InterceptedMessage),
		toBeDispatched: make(map[PeerID][]*InterceptedMessage),
		updateCh:       updateCh,
		lock:           new(sync.Mutex),
	}
}

func (m *msgStore) Add(msg *InterceptedMessage) {
	m.lock.Lock()
	m.messages[msg.ID] = msg
	m.lock.Unlock()
}

func (m *msgStore) Mark(msgID string) {
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
			m.toBeDispatched[msg.To] = make([]*InterceptedMessage, 0)
		}
		m.toBeDispatched[msg.To] = append(m.toBeDispatched[msg.To], msg)
		go func() {
			m.updateCh <- msg.To
		}()
	}
}

func (m *msgStore) Count(p PeerID) int {
	m.lock.Lock()
	defer m.lock.Unlock()

	count, ok := m.counters[p]
	if !ok {
		return 0
	}
	return count
}

func (m *msgStore) Fetch(p PeerID, limit int) ([]*InterceptedMessage, int) {
	m.lock.Lock()
	defer m.lock.Unlock()

	count, ok := m.counters[p]
	if !ok {
		return make([]*InterceptedMessage, 0), 0
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

func (m *msgStore) FetchOne(p PeerID) (*InterceptedMessage, error) {
	res, _ := m.Fetch(p, 1)
	if len(res) == 0 {
		return nil, errors.New("no messages for peer")
	}
	return res[0], nil
}

func (m *msgStore) Exists(msgID string) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	_, ok := m.messages[msgID]
	return ok
}

func (m *msgStore) Reset() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.counters = make(map[PeerID]int)
	m.toBeDispatched = make(map[PeerID][]*InterceptedMessage)
	m.messages = make(map[string]*InterceptedMessage)
}

type PeerStore struct {
	peers map[PeerID]*Peer
	lock  *sync.Mutex
}

func NewPeerStore() *PeerStore {
	return &PeerStore{
		peers: make(map[PeerID]*Peer),
		lock:  new(sync.Mutex),
	}
}

func (s *PeerStore) AddPeer(p *Peer) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.peers[p.ID] = p
}

func (s *PeerStore) GetPeer(id PeerID) (p *Peer, ok bool) {
	s.lock.Lock()
	p, ok = s.peers[id]
	s.lock.Unlock()
	return
}

func (s *PeerStore) Count() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return len(s.peers)
}

func (s *PeerStore) Iter() []*Peer {
	s.lock.Lock()
	defer s.lock.Unlock()
	peers := make([]*Peer, len(s.peers))
	i := 0
	for _, p := range s.peers {
		peers[i] = p
		i++
	}
	return peers
}
