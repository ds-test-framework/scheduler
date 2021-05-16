package types

import "sync"

type StateUpdatesStore struct {
	curRun  int
	updates map[int][]string

	lock *sync.Mutex
}

func NewStateUpdatesStore() *StateUpdatesStore {
	return &StateUpdatesStore{
		curRun:  0,
		updates: make(map[int][]string),
		lock:    new(sync.Mutex),
	}
}

func (s *StateUpdatesStore) SetRun(i int) {
	s.lock.Lock()
	s.curRun = i
	s.updates[s.curRun] = make([]string, 0)
	s.lock.Unlock()
}

func (s *StateUpdatesStore) AddUpdate(state string) {
	s.lock.Lock()
	s.updates[s.curRun] = append(s.updates[s.curRun], state)
	s.lock.Unlock()
}

type LogStore struct {
	curRun int
	lock   *sync.Mutex
	logs   map[int][]map[string]interface{}
}

func NewLogStore() *LogStore {
	return &LogStore{
		curRun: 0,
		logs:   make(map[int][]map[string]interface{}),
		lock:   new(sync.Mutex),
	}
}

func (s *LogStore) SetRun(i int) {
	s.lock.Lock()
	s.curRun = i
	s.logs[s.curRun] = make([]map[string]interface{}, 0)
	s.lock.Unlock()
}

func (s *LogStore) AddUpdate(params map[string]interface{}) {
	s.lock.Lock()
	s.logs[s.curRun] = append(s.logs[s.curRun], params)
	s.lock.Unlock()
}

type ReplicaStore struct {
	replicas map[ReplicaID]*Replica
	lock     *sync.Mutex
}

func NewReplicaStore() *ReplicaStore {
	return &ReplicaStore{
		replicas: make(map[ReplicaID]*Replica),
		lock:     new(sync.Mutex),
	}
}

func (s *ReplicaStore) AddReplica(p *Replica) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.replicas[p.ID] = p
}

func (s *ReplicaStore) GetReplica(id ReplicaID) (p *Replica, ok bool) {
	s.lock.Lock()
	p, ok = s.replicas[id]
	s.lock.Unlock()
	return
}

func (s *ReplicaStore) Count() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return len(s.replicas)
}

func (s *ReplicaStore) Iter() []*Replica {
	s.lock.Lock()
	defer s.lock.Unlock()
	replicas := make([]*Replica, len(s.replicas))
	i := 0
	for _, p := range s.replicas {
		replicas[i] = p
		i++
	}
	return replicas
}
