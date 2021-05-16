package types

import (
	"sync"

	"github.com/ds-test-framework/scheduler/pkg/log"
	"github.com/spf13/viper"
)

type ContextEvent struct {
	Type    ContextEventType
	Message *MessageWrapper
	Replica *Replica
}

func (c ContextEvent) Clone() ContextEvent {
	return ContextEvent{
		Type:    c.Type,
		Message: c.Message.Clone(),
		Replica: c.Replica.Clone(),
	}
}

type ContextEventType string

const (
	InterceptedMessage ContextEventType = "InterceptedMessage"
	NewReplica         ContextEventType = "NewReplica"
	EnabledMessage     ContextEventType = "EnabledMessage"
	ScheduledMessage   ContextEventType = "ScheduledMessage"
)

func (e ContextEventType) String() string {
	return string(e)
}

// Should have all the channels and subscribers not the actual instances of driver, engine
// Should not run its own go routines, Just outputting on channels
type Context struct {
	Replicas     *ReplicaStore
	StateUpdates *StateUpdatesStore
	Logs         *LogStore

	subscribers     map[ContextEventType][]chan ContextEvent
	subscribersLock *sync.Mutex

	run     int
	runLock *sync.Mutex

	config *viper.Viper

	Logger *log.Logger
}

func NewContext(config *viper.Viper, logger *log.Logger) *Context {
	return &Context{
		config:          config,
		StateUpdates:    NewStateUpdatesStore(),
		Logs:            NewLogStore(),
		Replicas:        NewReplicaStore(),
		subscribers:     make(map[ContextEventType][]chan ContextEvent),
		subscribersLock: new(sync.Mutex),

		run:     0,
		runLock: new(sync.Mutex),
		Logger:  logger,
	}
}

func (c *Context) Config(sub string) *viper.Viper {
	return c.config.Sub(sub)
}

func (c *Context) SetRun(run int) {
	c.runLock.Lock()
	c.run = run
	c.runLock.Unlock()
	c.StateUpdates.SetRun(run)
	c.Logs.SetRun(run)
}

func (c *Context) Subscribe(event ContextEventType) chan ContextEvent {
	c.subscribersLock.Lock()
	defer c.subscribersLock.Unlock()
	_, ok := c.subscribers[event]
	if !ok {
		c.subscribers[event] = make([]chan ContextEvent, 0)
	}

	outChan := make(chan ContextEvent, 10)
	c.subscribers[event] = append(c.subscribers[event], outChan)
	return outChan
}

func (c *Context) NewMessage(msg *Message) {
	c.runLock.Lock()
	run := c.run
	c.runLock.Unlock()

	msgW := &MessageWrapper{
		Run: run,
		Msg: msg,
	}

	c.dispatchEvent(ContextEvent{
		Type:    InterceptedMessage,
		Message: msgW,
	})
}

func (c *Context) NewReplica(r *Replica) {
	c.Replicas.AddReplica(r)

	go c.dispatchEvent(ContextEvent{
		Type:    NewReplica,
		Replica: r,
	})
}

func (c *Context) MarkMessage(msg *MessageWrapper) {
	c.runLock.Lock()
	run := c.run
	c.runLock.Unlock()

	if msg.Run != run {
		return
	}

	c.dispatchEvent(ContextEvent{
		Type:    EnabledMessage,
		Message: msg,
	})
}

func (c *Context) ScheduleMessage(msg *MessageWrapper) {
	c.runLock.Lock()
	run := c.run
	c.runLock.Unlock()

	if msg.Run != run {
		return
	}

	c.dispatchEvent(ContextEvent{
		Type:    ScheduledMessage,
		Message: msg,
	})
}

func (c *Context) dispatchEvent(e ContextEvent) {

	c.Logger.With(map[string]string{
		"event_type": e.Type.String(),
	}).Debug("Dispatching event")

	c.subscribersLock.Lock()
	chans, ok := c.subscribers[e.Type]
	if !ok {
		return
	}
	c.subscribersLock.Unlock()

	for _, c := range chans {
		go func(e ContextEvent, ch chan ContextEvent) {
			ch <- e
		}(e.Clone(), c)
	}
}
