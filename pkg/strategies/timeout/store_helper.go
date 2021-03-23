package timeout

import (
	"sync"

	"github.com/zeu5/model-checker/pkg/types"
)

type eventPseudoStore struct {
	events *eventStore
	dirty  map[uint]*types.Event
	lock   *sync.Mutex
}

func (s *eventPseudoStore) MarkDirty(eid uint) {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, ok := s.dirty[eid]
	if ok {
		return
	}

	event, ok := s.events.Get(eid)
	if !ok {
		// logger.Debug(fmt.Sprintf("Pseudo store: event does not exist: %d", eid))
		s.dirty[eid] = nil
		return
	}
	s.dirty[eid] = event.Clone()
}

func (s *eventPseudoStore) Get(eid uint) (*types.Event, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	event, ok := s.dirty[eid]
	if !ok {
		return s.events.Get(eid)
	}
	return event, true
}

func (s *eventPseudoStore) Set(e *types.Event) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.dirty[e.ID] = e
}

type messagePseudoStore struct {
	messages *messageStore
	dirty    map[string]*types.Message
	lock     *sync.Mutex
}

func (s *messagePseudoStore) MarkDirty(mid string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, ok := s.dirty[mid]
	if ok {
		return
	}

	message, ok := s.messages.Get(mid)
	if !ok {
		// logger.Debug(fmt.Sprintf("Pseudo store: message does not exist: %s", mid))
		s.dirty[mid] = nil
		return
	}
	s.dirty[mid] = message.Clone()
}

func (s *messagePseudoStore) Get(mid string) (*types.Message, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	message, ok := s.dirty[mid]
	if !ok {
		return s.messages.Get(mid)
	}
	return message, true
}

func (s *messagePseudoStore) Set(m *types.Message) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.dirty[m.ID] = m
}

type eventStore struct {
	store map[uint]*types.Event
	lock  *sync.Mutex
}

func newEventStore() *eventStore {
	return &eventStore{
		store: make(map[uint]*types.Event),
		lock:  new(sync.Mutex),
	}
}

func (s *eventStore) Get(id uint) (*types.Event, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	e, ok := s.store[id]
	if !ok {
		return nil, false
	}
	return e, true
}

func (s *eventStore) Set(e *types.Event) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.store[e.ID] = e
}

func (s *eventStore) Reset() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.store = make(map[uint]*types.Event)
}

func (s *eventStore) Pseudo() *eventPseudoStore {
	s.lock.Lock()
	defer s.lock.Unlock()

	return &eventPseudoStore{
		events: s,
		dirty:  make(map[uint]*types.Event),
		lock:   new(sync.Mutex),
	}
}

type messageStore struct {
	store map[string]*types.Message
	lock  *sync.Mutex
}

func newMessageStore() *messageStore {
	return &messageStore{
		store: make(map[string]*types.Message),
		lock:  new(sync.Mutex),
	}
}

func (s *messageStore) Reset() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.store = make(map[string]*types.Message)
}

func (s *messageStore) Get(id string) (*types.Message, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	e, ok := s.store[id]
	if !ok {
		return nil, false
	}
	return e, true
}

func (s *messageStore) Set(e *types.Message) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.store[e.ID] = e
}

func (s *messageStore) Pseudo() *messagePseudoStore {
	s.lock.Lock()
	defer s.lock.Unlock()

	return &messagePseudoStore{
		messages: s,
		dirty:    make(map[string]*types.Message),
		lock:     new(sync.Mutex),
	}
}
