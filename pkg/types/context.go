package types

import (
	"sync"

	"github.com/spf13/viper"
)

type ContextEvent struct {
	Type    ContextEventType
	Run     int
	Message *Message
	Replica *Replica
}

func (c ContextEvent) Clone() ContextEvent {
	return ContextEvent{
		Type:    c.Type,
		Run:     c.Run,
		Message: c.Message.Clone(),
		Replica: c.Replica.Clone(),
	}
}

type ContextEventType string

const (
	InterceptedMessage ContextEventType = "InterceptedMessage"
	NewReplica         ContextEventType = "NewReplica"
	EnabledMessage     ContextEventType = "EnabledMessage"
)

// Should have all the channels and subscribers not the actual instances of driver, engine
type Context struct {
	fromEngine chan *MessageWrapper
	toEngine   chan *MessageWrapper

	messages chan *Message

	Replicas     *ReplicaStore
	StateUpdates *RunStore
	Logs         *RunStore

	subscribers map[ContextEventType][]chan ContextEvent

	run     int
	runLock *sync.Mutex

	Config *viper.Viper
}

func NewContext(config *viper.Viper) *Context {
	return &Context{
		Config:       config,
		StateUpdates: NewRunStore(),
		Logs:         NewRunStore(),
		Replicas:     NewReplicaStore(),
		fromEngine:   make(chan *MessageWrapper, 10),
		toEngine:     make(chan *MessageWrapper, 10),

		messages:    make(chan *Message, 10),
		subscribers: make(map[ContextEventType][]chan ContextEvent),

		run:     0,
		runLock: new(sync.Mutex),
	}
}

func (c *Context) SetRun(run int) {
	c.runLock.Lock()
	c.run = run
	c.runLock.Unlock()
	c.StateUpdates.SetRun(run)
	c.Logs.SetRun(run)
}

func (c *Context) Subscribe(event ContextEventType) chan ContextEvent {
	_, ok := c.subscribers[event]
	if !ok {
		c.subscribers[event] = make([]chan ContextEvent, 0)
	}

	outChan := make(chan ContextEvent, 10)
	c.subscribers[event] = append(c.subscribers[event], outChan)
	return outChan
}

func (c *Context) NewMessage(msg *Message) {
	go func() {
		c.messages <- msg
	}()
}

func (c *Context) SendToEngine(msg *MessageWrapper) {
	go func() {
		c.toEngine <- msg
	}()
}

func (c *Context) NewReplica(r *Replica) {
	c.Replicas.AddReplica(r)

	go c.dispatchEvent(ContextEvent{
		Type:    NewReplica,
		Replica: r,
	})
}

func (c *Context) dispatchEvent(e ContextEvent) {
	chans, ok := c.subscribers[e.Type]
	if !ok {
		return
	}

	for _, c := range chans {
		go func(e ContextEvent, ch chan ContextEvent) {
			ch <- e
		}(e.Clone(), c)
	}
}
