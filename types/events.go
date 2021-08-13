package types

import (
	"sync"

	"github.com/ds-test-framework/scheduler/log"
)

type EventType interface {
	Clone() EventType
	Type() string
	String() string
}

type Event struct {
	Replica   ReplicaID `json:"replica"`
	Type      EventType `json:"-"`
	TypeS     string    `json:"type"`
	ID        uint64    `json:"id"`
	Timestamp int64     `json:"timestamp"`
}

func (e *Event) Clone() Clonable {
	return &Event{
		Replica:   e.Replica,
		Type:      e.Type.Clone(),
		TypeS:     e.TypeS,
		ID:        e.ID,
		Timestamp: e.Timestamp,
	}
}

type EventNodeSet struct {
	nodes map[uint64]*EventNode
	lock  *sync.Mutex
}

func NewEventNodeSet() *EventNodeSet {
	return &EventNodeSet{
		nodes: make(map[uint64]*EventNode),
		lock:  new(sync.Mutex),
	}
}

func (s *EventNodeSet) Add(n *EventNode) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.nodes[n.ID] = n
}

func (s *EventNodeSet) Get(id uint64) *EventNode {
	s.lock.Lock()
	n, ok := s.nodes[id]
	s.lock.Unlock()
	if !ok {
		return nil
	}
	return n
}

func (s *EventNodeSet) Iter() []*EventNode {
	result := make([]*EventNode, len(s.nodes))
	i := 0
	s.lock.Lock()
	for _, n := range s.nodes {
		result[i] = n
		i++
	}
	s.lock.Unlock()
	return result
}

func (s *EventNodeSet) Union(n *EventNodeSet) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, node := range n.Iter() {
		s.nodes[node.ID] = node
	}
}

func (s *EventNodeSet) Has(id uint64) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, ok := s.nodes[id]
	return ok
}

// EventQueue datastructure to store the messages in a FIFO queue
type EventQueue struct {
	events      []*Event
	subscribers map[string]chan *Event
	lock        *sync.Mutex
	size        int
	dispatchWG  *sync.WaitGroup
	*BaseService
}

// NewEventQueue returens an emtpy EventQueue
func NewEventQueue(logger *log.Logger) *EventQueue {
	return &EventQueue{
		events:      make([]*Event, 0),
		size:        0,
		subscribers: make(map[string]chan *Event),
		lock:        new(sync.Mutex),
		dispatchWG:  new(sync.WaitGroup),
		BaseService: NewBaseService("EventQueue", logger),
	}
}

// Start implements Service
func (q *EventQueue) Start() error {
	q.StartRunning()
	go q.dispatchloop()
	return nil
}

func (q *EventQueue) dispatchloop() {
	for {
		q.lock.Lock()
		size := q.size
		messages := q.events
		q.lock.Unlock()

		if size > 0 {
			toAdd := messages[0]

			for _, s := range q.subscribers {
				q.dispatchWG.Add(1)
				go func(subs chan *Event) {
					select {
					case subs <- toAdd.Clone().(*Event):
					case <-q.QuitCh():
					}
					q.dispatchWG.Done()
				}(s)
			}

			q.lock.Lock()
			q.size = q.size - 1
			q.events = q.events[1:]
			q.lock.Unlock()
		}
	}
}

// Stop implements Service
func (q *EventQueue) Stop() error {
	q.StopRunning()
	q.dispatchWG.Wait()
	return nil
}

// Add adds a message to the queue
func (q *EventQueue) Add(m *Event) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.events = append(q.events, m)
	q.size = q.size + 1
}

// Flush clears the queue of all messages
func (q *EventQueue) Flush() {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.events = make([]*Event, 0)
	q.size = 0
}

// Restart implements Service
func (q *EventQueue) Restart() error {
	q.Flush()
	return nil
}

func (q *EventQueue) Subscribe(label string) (chan *Event, error) {
	q.lock.Lock()
	defer q.lock.Unlock()
	_, ok := q.subscribers[label]
	if ok {
		return nil, ErrDuplicateSubs
	}
	newChan := make(chan *Event, 10)
	q.subscribers[label] = newChan
	return newChan, nil
}
