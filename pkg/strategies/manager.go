package strategies

import (
	"sync"

	"github.com/ds-test-framework/model-checker/pkg/types"
)

type EngineManager struct {
	engine types.StrategyEngine
	// Messages coming from the engine
	engineOut chan *types.MessageWrapper
	// Messages going to the engine
	engineIn chan *types.MessageWrapper
	// Messages from outside
	in chan *types.MessageWrapper
	// Messages to outside
	out      chan *types.MessageWrapper
	stopChan chan bool
	flush    bool
	run      int
	lock     *sync.Mutex
}

func NewEngineManager(
	engine types.StrategyEngine,
	inChan chan *types.MessageWrapper,
	outChan chan *types.MessageWrapper,
) *EngineManager {
	m := &EngineManager{
		engine:    engine,
		in:        inChan,
		out:       outChan,
		engineOut: make(chan *types.MessageWrapper, 10),
		engineIn:  make(chan *types.MessageWrapper, 10),
		stopChan:  make(chan bool, 2),
		flush:     false,
		lock:      new(sync.Mutex),
		run:       0,
	}

	m.engine.SetChannels(m.engineIn, m.engineOut)
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
		case msg, ok := <-m.in:
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
			// logger.Debug(fmt.Sprintf("Sending message to engine: %v, %#v", flush, msg.Msg))
			if !flush && msg.Run == run {
				m.engineIn <- msg
			}
		case msg, ok := <-m.engineOut:
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
				m.out <- msg
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
		l := len(m.in)
		if l == 0 {
			break
		}
		<-m.in
		// logger.Debug(fmt.Sprintf("Flushing: %#v", m))
	}

	for {
		l := len(m.engineOut)
		if l == 0 {
			break
		}
		<-m.engineOut
	}

	m.lock.Lock()
	m.flush = false
	m.lock.Unlock()
}

func (m *EngineManager) Stop() {
	m.stopChan <- true
}
