package types

import (
	"sync"
	"time"
)

type ReplicaLog struct {
	Replica   ReplicaID              `json:"id"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Params    map[string]interface{} `json:"params"`
}

func (l *ReplicaLog) Clone() Clonable {
	if l == nil {
		return nil
	}
	return &ReplicaLog{
		Replica: l.Replica,
		Params:  l.Params,
	}
}

type LogStore struct {
	curRun int
	lock   *sync.Mutex
	logs   map[int]map[ReplicaID][]*ReplicaLog
}

func NewLogStore() *LogStore {
	return &LogStore{
		curRun: 0,
		logs:   make(map[int]map[ReplicaID][]*ReplicaLog),
		lock:   new(sync.Mutex),
	}
}

func (s *LogStore) SetRun(i int) {
	s.lock.Lock()
	s.curRun = i
	s.logs[s.curRun] = make(map[ReplicaID][]*ReplicaLog)
	s.lock.Unlock()
}

func (s *LogStore) AddUpdate(l *ReplicaLog) {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, ok := s.logs[s.curRun]
	if !ok {
		s.logs[s.curRun] = make(map[ReplicaID][]*ReplicaLog)
	}

	_, ok = s.logs[s.curRun][l.Replica]
	if !ok {
		s.logs[s.curRun][l.Replica] = make([]*ReplicaLog, 0)
	}
	s.logs[s.curRun][l.Replica] = append(s.logs[s.curRun][l.Replica], l)
}

type ReplicaStore struct {
	replicas map[ReplicaID]*Replica
	size     int
	lock     *sync.Mutex
}

func NewReplicaStore(size int) *ReplicaStore {
	return &ReplicaStore{
		replicas: make(map[ReplicaID]*Replica),
		lock:     new(sync.Mutex),
		size:     size,
	}
}

func (s *ReplicaStore) Size() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.size
}

func (s *ReplicaStore) UpdateReplica(p *Replica) {
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

func (s *ReplicaStore) NumReady() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	count := 0
	for _, r := range s.replicas {
		if r.Ready {
			count = count + 1
		}
	}
	return count
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

func (s *ReplicaStore) ResetReady() {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, p := range s.replicas {
		p.Ready = false
	}
}
