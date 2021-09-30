package types

import (
	"encoding/json"
	"sync"

	"github.com/ds-test-framework/scheduler/log"
)

// EventType abstract type for representing different types of events
type EventType interface {
	// Clone copies the event type
	Clone() EventType
	// Type is a unique key for that event type
	Type() string
	// String should return a string representation of the event type
	String() string
}

// Event is a generic event that occurs at a replica
type Event struct {
	// Replica at which the event occurs
	Replica ReplicaID `json:"replica"`
	// Type of the event
	Type EventType `json:"-"`
	// TypeS is the string representation of the event
	TypeS string `json:"type"`
	// ID unique identifier assigned for every new event
	ID uint64 `json:"id"`
	// Timestamp of the event
	Timestamp int64 `json:"timestamp"`
}

func (e *Event) MessageID() (string, bool) {
	switch eType := e.Type.(type) {
	case *MessageReceiveEventType:
		return eType.MessageID, true
	case *MessageSendEventType:
		return eType.MessageID, true
	}
	return "", false
}

func (e *Event) IsMessageSend() bool {
	switch e.Type.(type) {
	case *MessageSendEventType:
		return true
	}
	return false
}

func (e *Event) IsMessageReceive() bool {
	switch e.Type.(type) {
	case *MessageReceiveEventType:
		return true
	}
	return false
}

// Clone implements Clonable
func (e *Event) Clone() Clonable {
	return &Event{
		Replica:   e.Replica,
		Type:      e.Type.Clone(),
		TypeS:     e.TypeS,
		ID:        e.ID,
		Timestamp: e.Timestamp,
	}
}

// EventNodeSet implements the set data structure of EventNode(s)
type EventNodeSet struct {
	nodes map[uint64]*EventNode
	lock  *sync.Mutex
}

// NewEventNodeSet instantiates EventNodeSet
func NewEventNodeSet() *EventNodeSet {
	return &EventNodeSet{
		nodes: make(map[uint64]*EventNode),
		lock:  new(sync.Mutex),
	}
}

// Add adds to the EventNodeSet
func (s *EventNodeSet) Add(n *EventNode) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.nodes[n.Event.ID] = n
}

// Get returns the EVentNode if it exists and nil otherwise
func (s *EventNodeSet) Get(id uint64) *EventNode {
	s.lock.Lock()
	n, ok := s.nodes[id]
	s.lock.Unlock()
	if !ok {
		return nil
	}
	return n
}

// Iter returns a list of all event nodes
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

// Union computes the union of the two sets
func (s *EventNodeSet) Union(n *EventNodeSet) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, node := range n.Iter() {
		s.nodes[node.Event.ID] = node
	}
}

// Has returns true if the event exists in the set
func (s *EventNodeSet) Has(id uint64) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, ok := s.nodes[id]
	return ok
}

// Size returns the size of the set
func (s *EventNodeSet) Size() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return len(s.nodes)
}

func (s *EventNodeSet) MarshalJSON() ([]byte, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	ids := make([]uint64, len(s.nodes))
	i := 0
	for id := range s.nodes {
		ids[i] = id
		i++
	}
	return json.Marshal(ids)
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
		subcribers := q.subscribers
		q.lock.Unlock()

		if size > 0 {
			toAdd := messages[0]

			for _, s := range subcribers {
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

// Subscribe creates and returns a channel for the subscriber with the given label
func (q *EventQueue) Subscribe(label string) chan *Event {
	q.lock.Lock()
	defer q.lock.Unlock()
	ch, ok := q.subscribers[label]
	if ok {
		return ch
	}
	newChan := make(chan *Event, 10)
	q.subscribers[label] = newChan
	return newChan
}
