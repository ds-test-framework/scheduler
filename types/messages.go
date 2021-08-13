package types

import (
	"errors"
	"sync"

	"github.com/ds-test-framework/scheduler/log"
)

var (
	ErrDuplicateSubs = errors.New("duplicate subscriber")
	ErrNoSubs        = errors.New("subscriber does not exist")

	DefaultSubsChSize = 10
)

// Message stores a message that has been interecepted between two replicas
type Message struct {
	From      ReplicaID `json:"from"`
	To        ReplicaID `json:"to"`
	Data      []byte    `json:"data"`
	Type      string    `json:"type"`
	ID        string    `json:"id"`
	Intercept bool      `json:"intercept"`
}

// Clone to create a new Message object with the same attributes
func (m *Message) Clone() Clonable {
	return &Message{
		From:      m.From,
		To:        m.To,
		Data:      m.Data,
		Type:      m.Type,
		ID:        m.ID,
		Intercept: m.Intercept,
	}
}

// MessageStore to store the messages. Thread safe
type MessageStore struct {
	messages map[string]*Message
	lock     *sync.Mutex
}

// NewMessageStore creates a empty MessageStore
func NewMessageStore() *MessageStore {
	return &MessageStore{
		messages: make(map[string]*Message),
		lock:     new(sync.Mutex),
	}
}

// Get returns a message and bool indicating if the message exists
func (s *MessageStore) Get(id string) (*Message, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	m, ok := s.messages[id]
	return m, ok
}

// Exists returns true if the message exists
func (s *MessageStore) Exists(id string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, ok := s.messages[id]
	return ok
}

// Add adds a message to the store
// Returns any old message with the same ID if it exists or nil if not
func (s *MessageStore) Add(m *Message) *Message {
	s.lock.Lock()
	defer s.lock.Unlock()

	old, ok := s.messages[m.ID]
	s.messages[m.ID] = m
	if ok {
		return old
	}
	return nil
}

// MessageQueue datastructure to store the messages in a FIFO queue
type MessageQueue struct {
	messages    []*Message
	subscribers map[string]chan *Message
	lock        *sync.Mutex
	size        int
	dispatchWG  *sync.WaitGroup
	*BaseService
}

// NewMessageQueue returens an emtpy MessageQueue
func NewMessageQueue(logger *log.Logger) *MessageQueue {
	return &MessageQueue{
		messages:    make([]*Message, 0),
		size:        0,
		subscribers: make(map[string]chan *Message),
		lock:        new(sync.Mutex),
		dispatchWG:  new(sync.WaitGroup),
		BaseService: NewBaseService("MessageQueue", logger),
	}
}

// Start implements Service
func (q *MessageQueue) Start() error {
	q.StartRunning()
	go q.dispatchloop()
	return nil
}

func (q *MessageQueue) dispatchloop() {
	for {
		q.lock.Lock()
		size := q.size
		messages := q.messages
		q.lock.Unlock()

		if size > 0 {
			toAdd := messages[0]

			for _, s := range q.subscribers {
				q.dispatchWG.Add(1)
				go func(subs chan *Message) {
					select {
					case subs <- toAdd.Clone().(*Message):
					case <-q.QuitCh():
					}
					q.dispatchWG.Done()
				}(s)
			}

			q.lock.Lock()
			q.size = q.size - 1
			q.messages = q.messages[1:]
			q.lock.Unlock()
		}
	}
}

// Stop implements Service
func (q *MessageQueue) Stop() error {
	q.StopRunning()
	q.dispatchWG.Wait()
	return nil
}

// Add adds a message to the queue
func (q *MessageQueue) Add(m *Message) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.messages = append(q.messages, m)
	q.size = q.size + 1
}

// Flush clears the queue of all messages
func (q *MessageQueue) Flush() {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.messages = make([]*Message, 0)
	q.size = 0
}

// Restart implements Service
func (q *MessageQueue) Restart() error {
	q.Flush()
	return nil
}

func (q *MessageQueue) Subscribe(label string) (chan *Message, error) {
	q.lock.Lock()
	defer q.lock.Unlock()
	_, ok := q.subscribers[label]
	if ok {
		return nil, ErrDuplicateSubs
	}
	newChan := make(chan *Message, DefaultSubsChSize)
	q.subscribers[label] = newChan
	return newChan, nil
}