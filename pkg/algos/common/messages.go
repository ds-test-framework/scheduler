package common

import (
	"encoding/json"

	"github.com/ds-test-framework/scheduler/pkg/types"
)

const (
	ErrFailedToUnmarshal = "FAILED_TO_UNMARSHAL"
	requestMessage       = "InterceptedMessage"
	requestPeerRegister  = "RegisterPeer"
	timeoutMessage       = "TimeoutMessage"
)

// InterceptedMessage represents a message sent by the instrumentaion between replicas
type InterceptedMessage struct {
	From      types.ReplicaID `json:"from"`
	To        types.ReplicaID `json:"to"`
	Msg       []byte          `json:"msg"`
	T         string          `json:"type"`
	ID        string          `json:"id"`
	Intercept bool            `json:"intercept"`
}

// Type returns the message type
func (i *InterceptedMessage) Type() string {
	return i.T
}

// Marshal serializes the message
func (i *InterceptedMessage) Marshal() ([]byte, error) {
	return json.Marshal(i)
}

// Unmarshal de-serializes the message
func (i *InterceptedMessage) Unmarshal(b []byte) error {
	return json.Unmarshal(b, i)
}

type timeout struct {
	Type     string          `json:"type"`
	Duration int             `json:"duration"`
	Peer     types.ReplicaID `json:"peer"`
}

type request struct {
	Type    string             `json:"type"`
	Peer    *Replica           `json:"peer,omitempty"`
	Message InterceptedMessage `json:"message,omitempty"`
	Timeout *timeout           `json:"timeout,omitempty"`
}

type response struct {
	Status string `json:"status"`
	Err    string `json:"error"`
}

func unmarshal(msg string) (*request, *types.Error) {
	r := &request{}
	err := json.Unmarshal([]byte(msg), r)
	if err != nil {
		return nil, types.NewError(
			ErrFailedToUnmarshal,
			"Failed to unmarshal request",
		)
	}
	return r, nil
}

// DirectiveMessage represents an action that a given replica should perform
type DirectiveMessage struct {
	Action string `json:"action"`
}

var (
	RestartDirective = DirectiveMessage{Action: "RESTART"}
	IsReadyDirective = DirectiveMessage{Action: "ISREADY"}
)
