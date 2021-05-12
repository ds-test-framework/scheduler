package types

import "sync"

type RunStore struct {
	curRun  int
	updates map[int][]string

	lock *sync.Mutex
}

func NewRunStore() *RunStore {
	return &RunStore{
		curRun:  0,
		updates: make(map[int][]string),
		lock:    new(sync.Mutex),
	}
}

func (s *RunStore) SetRun(i int) {
	s.lock.Lock()
	s.curRun = i
	s.updates[s.curRun] = make([]string, 0)
	s.lock.Unlock()
}

func (s *RunStore) AddUpdate(state string) {
	s.lock.Lock()
	s.updates[s.curRun] = append(s.updates[s.curRun], state)
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
