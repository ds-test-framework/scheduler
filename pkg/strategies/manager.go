package strategies

import (
	"sync"

	"github.com/ds-test-framework/scheduler/pkg/types"
)

type EngineManager struct {
	engine types.StrategyEngine
	// Messages coming from the engine
	fromEngine chan *types.MessageWrapper
	// Messages going to the engine
	toEngine chan *types.MessageWrapper
	// Messages from outside
	fromDriver chan *types.MessageWrapper
	// Messages to outside
	toDriver chan *types.MessageWrapper
	stopChan chan bool
	flush    bool
	run      int
	lock     *sync.Mutex
}

func NewEngineManager(
	engine types.StrategyEngine,
	fromDriver chan *types.MessageWrapper,
	toDriver chan *types.MessageWrapper,
) *EngineManager {
	m := &EngineManager{
		engine:     engine,
		fromDriver: fromDriver,
		toDriver:   toDriver,
		fromEngine: make(chan *types.MessageWrapper, 10),
		toEngine:   make(chan *types.MessageWrapper, 10),
		stopChan:   make(chan bool, 2),
		flush:      false,
		lock:       new(sync.Mutex),
		run:        0,
	}

	m.engine.SetChannels(m.toEngine, m.fromEngine)
	return m
}

func (m *EngineManager) SetRun(no int) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.run = no
}

func (m *EngineManager) Run() *types.Error {
	for {
		select {
		case msg, ok := <-m.fromDriver:
			if !ok {
				return types.NewError(
					types.ErrChannelClosed,
					"In channel closed",
				)
			}
			m.lock.Lock()
			flush := m.flush
			run := m.run
			m.lock.Unlock()
			// logger.Debug(fmt.Sprintf("Sending message to engine: %v, %#v for run %d", flush, msg.Msg, run))
			if !flush && msg.Run == run {
				m.toEngine <- msg
			}
		case msg, ok := <-m.fromEngine:
			if !ok {
				return types.NewError(
					types.ErrChannelClosed,
					"Out channel closed",
				)
			}
			// logger.Debug(fmt.Sprintf("Received message from engine: %#v", msg.Msg))
			m.lock.Lock()
			flush := m.flush
			run := m.run
			m.lock.Unlock()
			if !flush && msg.Run == run {
				m.toDriver <- msg
			}
		case _ = <-m.stopChan:
			return nil
		}
	}
}

func (m *EngineManager) FlushChannels() {
	m.lock.Lock()
	m.flush = true
	m.lock.Unlock()

	for {
		l := len(m.fromDriver)
		if l == 0 {
			break
		}
		<-m.fromDriver
		// logger.Debug(fmt.Sprintf("Flushing: %#v", m))
	}

	for {
		l := len(m.fromEngine)
		if l == 0 {
			break
		}
		<-m.fromEngine
	}

	m.lock.Lock()
	m.flush = false
	m.lock.Unlock()
}

func (m *EngineManager) Stop() {
	m.stopChan <- true
}
