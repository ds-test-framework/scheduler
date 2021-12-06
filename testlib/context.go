package testlib

import (
	"fmt"
	"sync"

	"github.com/ds-test-framework/scheduler/context"
	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
	"github.com/ds-test-framework/scheduler/util"
)

// Context struct is passed to the calls of StateAction and Condition
// encapsulates all information needed by the StateAction and Condition to function
type Context struct {
	// MessagePool reference to an instance of the MessageStore
	MessagePool *types.MessageStore
	// Replicas reference to the replica store
	Replicas *types.ReplicaStore
	// EventDAG is the directed acyclic graph all prior events
	EventDAG *types.EventDAG
	// Vars is a generic key value store to facilate maintaining auxilliary information
	// during the execution of a testcase
	Vars          *Vars
	TimeoutDriver *TimeoutDriver

	counter  *util.Counter
	testcase *TestCase
	report   *TestCaseReport
	sends    map[string]*types.Event
	lock     *sync.Mutex
	once     *sync.Once
}

// NewContext instantiates a Context from the RootContext
func NewContext(c *context.RootContext, testcase *TestCase, report *TestCaseReport) *Context {
	return &Context{
		MessagePool:   c.MessageStore,
		Replicas:      c.Replicas,
		EventDAG:      types.NewEventDag(),
		Vars:          NewVarSet(),
		TimeoutDriver: NewTimeoutDriver(c.TimeoutStore),

		counter:  util.NewCounter(),
		testcase: testcase,
		report:   report,
		sends:    make(map[string]*types.Event),
		lock:     new(sync.Mutex),
		once:     new(sync.Once),
	}
}

// Logger returns the logger for the current testcase
func (c *Context) Logger() *log.Logger {
	return c.testcase.Logger
}

// Abort stops the execution of the testcase
func (c *Context) Abort() {
	c.testcase.Abort()
}

func (c *Context) EndTestCase() {
	c.testcase.End()
}

func (c *Context) AddReportLog(message string, params map[string]interface{}) {
	c.report.Log.AddMessage(message, params)
}

func (c *Context) NewMessage(cur *types.Message, data []byte) *types.Message {
	return &types.Message{
		From:      cur.From,
		To:        cur.To,
		Data:      data,
		Type:      cur.Type,
		ID:        fmt.Sprintf("%s_%s_change%d", cur.From, cur.To, c.counter.Next()),
		Intercept: cur.Intercept,
	}
}

func (c *Context) GetMessage(e *types.Event) (*types.Message, bool) {
	if !e.IsMessageSend() && !e.IsMessageReceive() {
		return nil, false
	}
	mID, _ := e.MessageID()
	return c.MessagePool.Get(mID)
}

func (c *Context) setEvent(e *types.Event) {
	c.lock.Lock()
	defer c.lock.Unlock()

	parents := make([]*types.Event, 0)
	switch e.Type.(type) {
	case *types.MessageReceiveEventType:
		eventType := e.Type.(*types.MessageReceiveEventType)
		send, ok := c.sends[eventType.MessageID]
		if ok {
			parents = append(parents, send)
		}
	case *types.MessageSendEventType:
		eventType := e.Type.(*types.MessageSendEventType)
		c.sends[eventType.MessageID] = e
	}
	c.EventDAG.AddNode(e, parents)
	c.TimeoutDriver.NewEvent(e)
}
