package random

import (
	"sync"

	"github.com/ds-test-framework/scheduler/pkg/types"
	"github.com/ds-test-framework/scheduler/pkg/util"
)

// RandomScheduler picks a random message from the current pool of messages
type RandomScheduler struct {
	inChan chan types.ContextEvent
	ctx    *types.Context
	stopCh chan bool

	msgMap map[string]*types.MessageWrapper
	lock   *sync.Mutex
}

// NewRandomScheduler returns a RandomScheduler
func NewRandomScheduler(ctx *types.Context) *RandomScheduler {
	return &RandomScheduler{
		stopCh: make(chan bool, 1),
		msgMap: make(map[string]*types.MessageWrapper),
		lock:   new(sync.Mutex),
	}
}

func (r *RandomScheduler) dispatch(msgID string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	m, ok := r.msgMap[msgID]
	if ok {
		delete(r.msgMap, msgID)
		r.ctx.MarkMessage(m)
	}
}

func (r *RandomScheduler) pickRandom() *types.MessageWrapper {
	r.lock.Lock()
	length := len(r.msgMap)
	keys := make([]string, length)
	i := 0
	for k := range r.msgMap {
		keys[i] = k
		i++
	}
	r.lock.Unlock()

	if length == 0 {
		return nil
	}

	randIndex := util.RandIntn(length)
	key := keys[randIndex]

	r.lock.Lock()
	m, ok := r.msgMap[key]
	r.lock.Unlock()

	if ok {
		return m
	}
	return nil
}

func (r *RandomScheduler) scheduleMessages() {
	for {
		select {
		case <-r.stopCh:
			return
		default:
		}

		m := r.pickRandom()
		if m != nil {
			go r.dispatch(m.Msg.ID)
		}
	}
}

func (r *RandomScheduler) pollInChan() {
	for {
		select {
		case e := <-r.inChan:
			m := e.Message
			r.lock.Lock()
			r.msgMap[m.Msg.ID] = m
			r.lock.Unlock()
		case <-r.stopCh:
			return
		}
	}
}

// Start implements StrategyEngine
func (r *RandomScheduler) Start() *types.Error {
	go r.pollInChan()
	go r.scheduleMessages()
	return nil
}

// Stop implements StrategyEngine
func (r *RandomScheduler) Stop() {
	close(r.stopCh)
}

// Reset implements StrategyEngine
func (r *RandomScheduler) Reset() {
	r.lock.Lock()
	r.msgMap = make(map[string]*types.MessageWrapper)
	r.lock.Unlock()
}
