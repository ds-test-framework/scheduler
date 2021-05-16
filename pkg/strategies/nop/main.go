package nop

import (
	"github.com/ds-test-framework/scheduler/pkg/types"
)

// NopScheduler does nothing. Just returns the incoming message in the outgoing channel
type NopScheduler struct {
	inChan chan types.ContextEvent
	stopCh chan bool
	ctx    *types.Context
}

// NewNopScheduler returns a new NopScheduler
func NewNopScheduler(ctx *types.Context) *NopScheduler {
	return &NopScheduler{
		stopCh: make(chan bool, 1),
		ctx:    ctx,
		inChan: ctx.Subscribe(types.ScheduledMessage),
	}
}

// Reset implements StrategyEngine
func (n *NopScheduler) Reset() {
}

// Run implements StrategyEngine
func (n *NopScheduler) Start() *types.Error {
	go n.poll()
	return nil
}

// Stop implements StrategyEngine
func (n *NopScheduler) Stop() {
	close(n.stopCh)
}

func (n *NopScheduler) poll() {
	for {
		select {
		case e := <-n.inChan:
			m := e.Message
			go func(m *types.MessageWrapper) {
				n.ctx.MarkMessage(m)
			}(m)
		case <-n.stopCh:
			return
		}
	}
}
