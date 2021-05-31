package util

import "sync"

type IDGenerator struct {
	counter int
	mtx     *sync.Mutex
}

func NewIDGenerator() *IDGenerator {
	return &IDGenerator{
		counter: 0,
		mtx:     new(sync.Mutex),
	}
}

func (id *IDGenerator) Next() int {
	id.mtx.Lock()
	defer id.mtx.Unlock()

	cur := id.counter
	id.counter = id.counter + 1

	return cur
}

func (id *IDGenerator) Reset() {
	id.mtx.Lock()
	defer id.mtx.Unlock()

	id.counter = 0
}
