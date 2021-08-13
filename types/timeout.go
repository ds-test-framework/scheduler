package types

type ReplicaTimeout struct {
	Replica  ReplicaID `json:"replica"`
	Type     string    `json:"type"`
	Duration int64     `json:"duration"`
}
