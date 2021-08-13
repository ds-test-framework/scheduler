package ttest

import (
	"sync"
	"time"

	"github.com/ds-test-framework/scheduler/types"
)

type delayStrategy interface {
	AddMessage(*types.MessageWrapper) error
	Out() chan *types.MessageWrapper
	Start() error
	Stop() error
	Reset() error
}

type timeDelayStrategy struct {
	duration time.Duration
	buffer   []*types.MessageWrapper
	outChan  chan *types.MessageWrapper
	updates  chan bool
	mtx      *sync.Mutex
	stopCh   chan bool
}

func newTimeDelayStrategy(duration time.Duration) *timeDelayStrategy {
	return &timeDelayStrategy{
		duration: duration,
		buffer:   make([]*types.MessageWrapper, 0),
		outChan:  make(chan *types.MessageWrapper, 10),
		updates:  make(chan bool, 10),
		stopCh:   make(chan bool),
		mtx:      new(sync.Mutex),
	}
}

func (t *timeDelayStrategy) AddMessage(msg *types.MessageWrapper) error {
	t.mtx.Lock()
	t.buffer = append(t.buffer, msg)
	t.mtx.Unlock()
	go func() {
		t.updates <- true
	}()

	return nil
}

func (t *timeDelayStrategy) Out() chan *types.MessageWrapper {
	return t.outChan
}

func (t *timeDelayStrategy) Start() error {
	go t.poll()
	return nil
}

func (t *timeDelayStrategy) Stop() error {
	close(t.stopCh)
	return nil
}

func (t *timeDelayStrategy) Reset() error {
	return nil
}

func (t *timeDelayStrategy) delay(msg *types.MessageWrapper) {
	select {
	case <-time.After(t.duration):
		t.outChan <- msg
	case <-t.stopCh:
		return
	}
}

func (t *timeDelayStrategy) poll() {
	for {
		select {
		case <-t.stopCh:
			return
		case <-t.updates:
			t.mtx.Lock()
			msg := t.buffer[0]
			t.buffer = t.buffer[1:len(t.buffer)]
			t.mtx.Unlock()
			t.delay(msg)
		}
	}
}
