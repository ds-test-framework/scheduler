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
	InterceptedMessage   ContextEventType = "InterceptedMessage"
	ReplicaUpdate        ContextEventType = "ReplicaUpdate"
	EnabledMessage       ContextEventType = "EnabledMessage"
	ScheduledMessage     ContextEventType = "ScheduledMessage"
	UnInterceptedMessage ContextEventType = "UnInterceptedMessage"
	LogMessage           ContextEventType = "LogMessage"
	EventMessage         ContextEventType = "EventMessage"
)

func (e ContextEventType) String() string {
	return string(e)
}

type Run struct {
	Id    int    `json:"id"`
	Label string `json:"label"`
}

// Should have all the channels and subscribers not the actual instances of driver, engine
// Should not run its own go routines, Just outputting on channels
type Context struct {
	Replicas   *ReplicaStore
	Logs       *LogStore
	IDGen      *util.IDGenerator
	EventGraph *GraphHelper
	FaultModel *FaultModel

	subscribers     map[ContextEventType][]chan ContextEvent
	subscribersLock *sync.Mutex

	run     int
	Runs    []*Run
	runLock *sync.Mutex

	config *viper.Viper

	Logger *log.Logger
}

func NewContext(config *viper.Viper, logger *log.Logger) *Context {
	config.SetDefault("driver.num_replicas", 4)
	replicaSize := config.Sub("driver").GetInt("num_replicas")

	return &Context{
		config:          config,
		Logs:            NewLogStore(),
		Replicas:        NewReplicaStore(replicaSize),
		EventGraph:      NewGraphHelper(),
		IDGen:           util.NewIDGenerator(),
		subscribers:     make(map[ContextEventType][]chan ContextEvent),
		subscribersLock: new(sync.Mutex),
		FaultModel:      NewFaultModel(config),

		run:     0,
		Runs:    make([]*Run, 0),
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

func (c *Context) SetRun(run *Run) {
	c.runLock.Lock()
	c.run = run.Id
	c.Runs = append(c.Runs, run)
	c.runLock.Unlock()
	c.Logs.SetRun(run.Id)
	c.EventGraph.SetRun(run.Id)
}

func (c *Context) GetCurRun() int {
	c.runLock.Lock()
	defer c.runLock.Unlock()
	return c.run
}

func (c *Context) GetAllRuns() []*Run {
	c.runLock.Lock()
	defer c.runLock.Unlock()
	return c.Runs
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

	c.Logger.With(map[string]interface{}{
		"event_type": t.String(),
	}).Debug("Dispatching event")

	event := ContextEvent{
		Data: data,
		Type: t,
	}

	switch t {
	case EventMessage:
		event := data.(*Event)
		c.EventGraph.AddEvent(event)
	}

	for _, c := range chans {
		go func(e ContextEvent, ch chan ContextEvent) {
			ch <- e
		}(event.Clone(), c)
	}
}

type GraphState struct {
	Graph           *EventGraph
	messageEventMap map[string][]*Event
	latestEvent     map[ReplicaID]*Event
	lock            *sync.Mutex
}

func NewGraphState() *GraphState {
	return &GraphState{
		Graph:           NewEventGraph(),
		messageEventMap: make(map[string][]*Event),
		latestEvent:     make(map[ReplicaID]*Event),
		lock:            new(sync.Mutex),
	}
}

func (s *GraphState) AddEvent(e *Event) {
	s.lock.Lock()
	defer s.lock.Unlock()

	parents := make([]*Event, 0)

	curLatest, ok := s.latestEvent[e.Replica]
	if ok {
		curLatest.UpdateNext(e)
		e.UpdatePrev(curLatest)
		parents = append(parents, curLatest)
	}
	s.latestEvent[e.Replica] = e

	if e.Type.Type() == SendMessageTypeS || e.Type.Type() == ReceiveMessageTypeS {
		msgID, _ := e.MessageID()
		messageEvents, ok := s.messageEventMap[msgID]
		if !ok {
			s.messageEventMap[msgID] = make([]*Event, 0)
		}
		if e.Type.Type() == ReceiveMessageTypeS {
			parents = append(parents, messageEvents...)
		}
		s.messageEventMap[msgID] = append(s.messageEventMap[msgID], e)
	}
	s.Graph.AddEvent(e, parents)
}

type GraphHelper struct {
	graphRunMap map[int]*GraphState
	curRun      int
	lock        *sync.Mutex
}

func NewGraphHelper() *GraphHelper {
	return &GraphHelper{
		graphRunMap: make(map[int]*GraphState),
		curRun:      0,
		lock:        new(sync.Mutex),
	}
}

func (h *GraphHelper) SetRun(run int) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.curRun = run
	_, ok := h.graphRunMap[run]
	if !ok {
		h.graphRunMap[run] = NewGraphState()
	}
}

func (h *GraphHelper) AddEvent(e *Event) {
	h.lock.Lock()
	state, ok := h.graphRunMap[h.curRun]
	h.lock.Unlock()

	if !ok {
		return
	}

	state.AddEvent(e)
}

func (h *GraphHelper) GetCurGraph() *EventGraph {
	h.lock.Lock()
	defer h.lock.Unlock()

	return h.graphRunMap[h.curRun].Graph
}

func (h *GraphHelper) GetGraph(run int) (*EventGraph, bool) {
	h.lock.Lock()
	defer h.lock.Unlock()

	graphState, ok := h.graphRunMap[run]
	if !ok {
		return nil, false
	}
	return graphState.Graph, true
}
