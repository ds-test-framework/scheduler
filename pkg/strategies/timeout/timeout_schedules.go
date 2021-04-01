package timeout

import (
	"sync"
	"time"

	"github.com/ds-test-framework/scheduler/pkg/types"
)

type scheduleManager struct {
	schedules map[uint]*schedule
	outChan   chan *types.Event
	lock      *sync.Mutex
}

func newScheduleManager(outChan chan *types.Event) *scheduleManager {
	return &scheduleManager{
		schedules: make(map[uint]*schedule),
		outChan:   outChan,
		lock:      new(sync.Mutex),
	}
}

func (m *scheduleManager) monitorSchedule(s *schedule) {
	s.Run()
	m.lock.Lock()
	_, ok := m.schedules[s.event.ID]
	if ok {
		delete(m.schedules, s.event.ID)
	}
	m.lock.Unlock()
}

func (m *scheduleManager) Schedule(e *types.Event, d time.Duration) {
	s := newSchedule(e, d, m.outChan)
	go m.monitorSchedule(s)
}

func (m *scheduleManager) Reset() {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, s := range m.schedules {
		s.stopChan <- true
	}
}

type schedule struct {
	stopChan chan bool
	event    *types.Event
	d        time.Duration
	outChan  chan *types.Event
}

func newSchedule(e *types.Event, d time.Duration, outChan chan *types.Event) *schedule {
	return &schedule{
		stopChan: make(chan bool, 2),
		event:    e,
		d:        d,
		outChan:  outChan,
	}
}

func (s *schedule) Run() {
	select {
	case <-time.After(s.d):
		s.outChan <- s.event
		return
	case <-s.stopChan:
		return
	}
}
