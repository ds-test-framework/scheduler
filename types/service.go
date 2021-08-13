package types

import (
	"sync"

	"github.com/ds-test-framework/scheduler/log"
)

type Service interface {
	Name() string
	Start() error
	Running() bool
	Stop() error
	Quit() <-chan struct{}
}

type RestartableService interface {
	Service
	Restart() error
}

type BaseService struct {
	running bool
	o       *sync.Once
	lock    *sync.Mutex
	name    string
	quit    chan struct{}
	Logger  *log.Logger
}

func NewBaseService(name string, parentLogger *log.Logger) *BaseService {
	return &BaseService{
		running: false,
		lock:    new(sync.Mutex),
		name:    name,
		o:       new(sync.Once),
		quit:    make(chan struct{}),
		Logger:  parentLogger.With(log.LogParams{"service": name}),
	}
}

func (b *BaseService) StartRunning() {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.running = true
}

func (b *BaseService) StopRunning() {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.running = false
	b.o.Do(func() {
		close(b.quit)
	})
}

func (b *BaseService) Name() string {
	return b.name
}

func (b *BaseService) Running() bool {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.running
}

func (b *BaseService) QuitCh() <-chan struct{} {
	return b.quit
}
