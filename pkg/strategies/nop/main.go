package nop

import (
	"github.com/ds-test-framework/scheduler/pkg/types"
)

// NopScheduler does nothing. Just returns the incoming message in the outgoing channel
type NopScheduler struct {
	inChan  chan *types.MessageWrapper
	outChan chan *types.MessageWrapper
	stopCh  chan bool
}

// NewNopScheduler returns a new NopScheduler
func NewNopScheduler() *NopScheduler {
	return &NopScheduler{
		stopCh: make(chan bool, 1),
	}
}

// Reset implements StrategyEngine
func (n *NopScheduler) Reset() {
}

// Run implements StrategyEngine
func (n *NopScheduler) Run() *types.Error {
	go n.poll()
	return nil
}

// Stop implements StrategyEngine
func (n *NopScheduler) Stop() {
	close(n.stopCh)
}

// SetChannels implements StrategyEngine
func (n *NopScheduler) SetChannels(inChan chan *types.MessageWrapper, outChan chan *types.MessageWrapper) {
	n.inChan = inChan
	n.outChan = outChan
}

func (n *NopScheduler) poll() {
	for {
		select {
		case m := <-n.inChan:
			go func(m *types.MessageWrapper) {
				n.outChan <- m
			}(m)
		case <-n.stopCh:
			return
		}
	}
}
