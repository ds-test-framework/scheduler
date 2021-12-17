package testlib

import (
	"sync"

	"github.com/ds-test-framework/scheduler/types"
)

// Vars is a dictionary for storing auxilliary state during the execution of the testcase
// Vars is stored in the context passed to actions and conditions
type Vars struct {
	vars map[string]interface{}
	lock *sync.Mutex
}

// NewVarSet instantiates Vars
func NewVarSet() *Vars {
	return &Vars{
		vars: make(map[string]interface{}),
		lock: new(sync.Mutex),
	}
}

// Get returns the value stored  of the specified label
// the second return argument is false if the label does not exist
func (v *Vars) Get(label string) (interface{}, bool) {
	v.lock.Lock()
	defer v.lock.Unlock()

	val, ok := v.vars[label]
	return val, ok
}

// GetInt casts the value at label (if it exists) into integer and returns it
func (v *Vars) GetInt(label string) (int, bool) {
	val, ok := v.Get(label)
	if !ok {
		return 0, ok
	}
	valInt, ok := val.(int)
	return valInt, ok
}

// GetString casts the value at label (if it exists) into string and returns it
func (v *Vars) GetString(label string) (string, bool) {
	val, ok := v.Get(label)
	if !ok {
		return "", ok
	}
	valS, ok := val.(string)
	return valS, ok
}

// GetBool casts the value at label (if it exists) into boolean and returns it
func (v *Vars) GetBool(label string) (bool, bool) {
	val, ok := v.Get(label)
	if !ok {
		return false, false
	}
	valB, ok := val.(bool)
	return valB, ok
}

// Set the value at the specified label
func (v *Vars) Set(label string, value interface{}) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.vars[label] = value
}

// Exists returns true if there is a variable of the specified key
func (v *Vars) Exists(label string) bool {
	v.lock.Lock()
	defer v.lock.Unlock()
	_, ok := v.vars[label]
	return ok
}

// SetCounter sets a counter instance at the specified label with initial value 1
func (v *Vars) SetCounter(label string) {
	v.lock.Lock()
	defer v.lock.Unlock()
	counter := NewCounter()
	counter.Incr()
	v.vars[label] = counter
}

// GetCounter returns the counter at the label if it exists (nil, false) otherwise
func (v *Vars) GetCounter(label string) (*Counter, bool) {
	v.lock.Lock()
	defer v.lock.Unlock()
	cI, exists := v.vars[label]
	if !exists {
		return nil, false
	}
	counter, ok := cI.(*Counter)
	return counter, ok
}

// NewMessageSet creates a message set at the specified label
func (v *Vars) NewMessageSet(label string) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.vars[label] = types.NewMessageStore()
}

// GetMessageSet returns the message set at label if one exists (nil, false) otherwise
func (v *Vars) GetMessageSet(label string) (*types.MessageStore, bool) {
	v.lock.Lock()
	defer v.lock.Unlock()
	sI, exists := v.vars[label]
	if !exists {
		return nil, false
	}
	set, ok := sI.(*types.MessageStore)
	return set, ok
}

// Counter threadsafe counter
type Counter struct {
	val  int
	lock *sync.Mutex
}

// NewCounter returns a counter
func NewCounter() *Counter {
	return &Counter{
		val:  0,
		lock: new(sync.Mutex),
	}
}

func (c *Counter) SetValue(v int) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.val = v
}

// Incr increments the counter
func (c *Counter) Incr() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.val = c.val + 1
}

// Value returns the counter value
func (c *Counter) Value() int {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.val
}
