package context

import (
	"github.com/ds-test-framework/scheduler/config"
	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
	"github.com/ds-test-framework/scheduler/util"
)

type RootContext struct {
	Config       *config.Config
	Replicas     *types.ReplicaStore
	LogStore     *types.ReplicaLogStore
	MessageQueue *types.MessageQueue
	MessageStore *types.MessageStore
	EventQueue   *types.EventQueue
	LogQueue     *types.ReplicaLogQueue
	Counter      *util.Counter
	Logger       *log.Logger
}

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

func (c *RootContext) Start() {
	c.MessageQueue.Start()
	c.EventQueue.Start()
	c.LogQueue.Start()
}

func (c *RootContext) Stop() {
	c.MessageQueue.Stop()
	c.EventQueue.Stop()
	c.LogQueue.Stop()
}
