package testing

import (
	"fmt"
	"sync"
)

type resultMap struct {
	success int
	total   int

	m map[string]bool

	mtx *sync.Mutex
}

func newResultMap() *resultMap {
	return &resultMap{
		success: 0,
		total:   0,
		m:       make(map[string]bool),
		mtx:     new(sync.Mutex),
	}
}

func (rm *resultMap) add(name string, ok bool) {
	rm.mtx.Lock()
	defer rm.mtx.Unlock()

	rm.total = rm.total + 1
	if ok {
		rm.success = rm.success + 1
	} else {
		rm.m[name] = ok
	}
}

func (rm *resultMap) shortSummary() (int, int) {
	rm.mtx.Lock()
	defer rm.mtx.Unlock()
	return rm.total, rm.success
}

func (rm *resultMap) summary() string {
	str := ""

	rm.mtx.Lock()
	defer rm.mtx.Unlock()

	str += fmt.Sprintf("Total: %d, Success: %d, Failed: %d\n", rm.total, rm.success, rm.total-rm.success)
	return str
}
