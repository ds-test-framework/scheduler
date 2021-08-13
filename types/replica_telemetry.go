package types

import "sync"

type ReplicaLog struct {
	Params    map[string]string `json:"params"`
	Message   string            `json:"message"`
	Timestamp int64             `json:"timestamp"`
	Replica   ReplicaID         `json:"replica"`
}

type ReplicaLogStore struct {
	logs map[ReplicaID][]*ReplicaLog
	lock *sync.Mutex
}

func NewReplicaLogStore() *ReplicaLogStore {
	return &ReplicaLogStore{
		logs: make(map[ReplicaID][]*ReplicaLog),
		lock: new(sync.Mutex),
	}
}

type ReplicaState struct {
	State     string    `json:"state"`
	Timestamp int64     `json:"timestamp"`
	Replica   ReplicaID `json:"replica"`
}

type ReplicaStateStore struct {
	stateUpdates map[ReplicaID]*ReplicaState
	lock         *sync.Mutex
}

func NewReplicaStateStore() *ReplicaStateStore {
	return &ReplicaStateStore{
		stateUpdates: make(map[ReplicaID]*ReplicaState),
		lock:         new(sync.Mutex),
	}
}
