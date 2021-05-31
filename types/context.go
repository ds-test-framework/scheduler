package types

import (
	"sync"

	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/util"
	"github.com/spf13/viper"
)

type ContextEvent struct {
	Data Clonable
	Type ContextEventType
}

func (c ContextEvent) Clone() ContextEvent {
	return ContextEvent{
		Type: c.Type,
		Data: c.Data.Clone(),
	}
}

type ContextEventType string

const (
	InterceptedMessage ContextEventType = "InterceptedMessage"
	ReplicaUpdate      ContextEventType = "ReplicaUpdate"
	EnabledMessage     ContextEventType = "EnabledMessage"
	ScheduledMessage   ContextEventType = "ScheduledMessage"
	LogMessage         ContextEventType = "LogMessage"
	StateMessage       ContextEventType = "StateMessage"
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
	IDGen        *util.IDGenerator

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
		IDGen:           util.NewIDGenerator(),
		subscribers:     make(map[ContextEventType][]chan ContextEvent),
		subscribersLock: new(sync.Mutex),

		run:     0,
		runLock: new(sync.Mutex),
		Logger:  logger,
	}
}

func (c *Context) Config(sub string) *viper.Viper {
	if sub == "root" {
		return c.config
	}
	return c.config.Sub(sub)
}

func (c *Context) SetRun(run int) {
	c.runLock.Lock()
	c.run = run
	c.runLock.Unlock()
	c.StateUpdates.SetRun(run)
	c.Logs.SetRun(run)
}

func (c *Context) GetRun() int {
	c.runLock.Lock()
	defer c.runLock.Unlock()
	return c.run
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

func (c *Context) Publish(t ContextEventType, data Clonable) {
	c.subscribersLock.Lock()
	chans, ok := c.subscribers[t]
	c.subscribersLock.Unlock()

	if !ok {
		return
	}

	c.Logger.With(map[string]string{
		"event_type": t.String(),
	}).Debug("Dispatching event")

	event := ContextEvent{
		Data: data,
		Type: t,
	}

	for _, c := range chans {
		go func(e ContextEvent, ch chan ContextEvent) {
			ch <- e
		}(event.Clone(), c)
	}
}
