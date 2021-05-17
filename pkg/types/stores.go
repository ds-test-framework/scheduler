package types

import "sync"

type StateUpdate struct {
	Replica ReplicaID `json:"id"`
	State   string    `json:"state"`
}

type StateUpdatesStore struct {
	curRun  int
	updates map[int]map[ReplicaID][]string

	lock *sync.Mutex
}

func NewStateUpdatesStore() *StateUpdatesStore {
	return &StateUpdatesStore{
		curRun:  0,
		updates: make(map[int]map[ReplicaID][]string),
		lock:    new(sync.Mutex),
	}
}

func (s *StateUpdatesStore) SetRun(i int) {
	s.lock.Lock()
	s.curRun = i
	s.updates[s.curRun] = make(map[ReplicaID][]string)
	s.lock.Unlock()
}

func (s *StateUpdatesStore) AddUpdate(update *StateUpdate) {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, ok := s.updates[s.curRun]
	if !ok {
		s.updates[s.curRun] = make(map[ReplicaID][]string)
	}

	_, ok = s.updates[s.curRun][update.Replica]
	if !ok {
		s.updates[s.curRun][update.Replica] = make([]string, 0)
	}
	s.updates[s.curRun][update.Replica] = append(s.updates[s.curRun][update.Replica], update.State)

}

type ReplicaLog struct {
	Replica ReplicaID              `json:"id"`
	Params  map[string]interface{} `json:"params"`
}

type LogStore struct {
	curRun int
	lock   *sync.Mutex
	logs   map[int]map[ReplicaID][]map[string]interface{}
}

func NewLogStore() *LogStore {
	return &LogStore{
		curRun: 0,
		logs:   make(map[int]map[ReplicaID][]map[string]interface{}),
		lock:   new(sync.Mutex),
	}
}

func (s *LogStore) SetRun(i int) {
	s.lock.Lock()
	s.curRun = i
	s.logs[s.curRun] = make(map[ReplicaID][]map[string]interface{})
	s.lock.Unlock()
}

func (s *LogStore) AddUpdate(l *ReplicaLog) {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, ok := s.logs[s.curRun]
	if !ok {
		s.logs[s.curRun] = make(map[ReplicaID][]map[string]interface{})
	}

	_, ok = s.logs[s.curRun][l.Replica]
	if !ok {
		s.logs[s.curRun][l.Replica] = make([]map[string]interface{}, 0)
	}
	s.logs[s.curRun][l.Replica] = append(s.logs[s.curRun][l.Replica], l.Params)
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
