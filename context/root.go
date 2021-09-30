package context

import (
	"github.com/ds-test-framework/scheduler/config"
	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
	"github.com/ds-test-framework/scheduler/util"
)

// RootContext stores the context of the scheduler
type RootContext struct {
	// Config and instance of the configuration object
	Config *config.Config
	// Replicas instance of the ReplicaStore which contains information of all the replicas
	Replicas *types.ReplicaStore
	// LogStore which stores the log messages sent by the replicas
	LogStore *types.ReplicaLogStore
	// MessageQueue stores the messages that are _intercepted_ as a queue
	MessageQueue *types.MessageQueue
	// MessageStore stores the messages that are _intercepted_ as a map
	MessageStore *types.MessageStore
	// EventQueue stores the events sent by the replicas as a queue
	EventQueue *types.EventQueue
	// LogQueue stores the log messages sent by the replicas as a queue
	LogQueue *types.ReplicaLogQueue
	// Counter is a thread safe monotonic integer counter
	Counter *util.Counter
	// Logger for logging purposes
	Logger *log.Logger
}

// NewRootContext creates an instance of the RootContext from the configuration
func NewRootContext(config *config.Config, logger *log.Logger) *RootContext {
	return &RootContext{
		Config:       config,
		Replicas:     types.NewReplicaStore(config.NumReplicas),
		LogStore:     types.NewReplicaLogStore(),
		MessageQueue: types.NewMessageQueue(logger),
		MessageStore: types.NewMessageStore(),
		EventQueue:   types.NewEventQueue(logger),
		LogQueue:     types.NewReplicaLogQueue(logger),
		Counter:      util.NewCounter(),
		Logger:       logger,
	}
}

// Start implements Service and initializes the queues
func (c *RootContext) Start() {
	c.MessageQueue.Start()
	c.EventQueue.Start()
	c.LogQueue.Start()
}

// Stop implements Service and terminates the queues
func (c *RootContext) Stop() {
	c.MessageQueue.Stop()
	c.EventQueue.Stop()
	c.LogQueue.Stop()
}

func (c *RootContext) Reset() {
	c.MessageQueue.Flush()
	c.EventQueue.Flush()
	c.LogQueue.Flush()

	c.MessageStore.RemoveAll()
	c.LogStore.Reset()
}
