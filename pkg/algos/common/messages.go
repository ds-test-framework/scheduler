package common

import (
	"encoding/json"

	"github.com/ds-test-framework/scheduler/pkg/types"
)

const (
	ErrFailedToUnmarshal = "FAILED_TO_UNMARSHAL"
)

var (
	RequestMessage      string = "InterceptedMessage"
	RequestPeerRegister string = "RegisterPeer"
	TimeoutMessage      string = "TimeoutMessage"
)

type InterceptedMessage struct {
	From      PeerID `json:"from"`
	To        PeerID `json:"to"`
	Msg       []byte `json:"msg"`
	T         string `json:"type"`
	ID        string `json:"id"`
	Intercept bool   `json:"intercept"`
}

func (i *InterceptedMessage) Type() string {
	return i.T
}

func (i *InterceptedMessage) Marshal() ([]byte, error) {
	return json.Marshal(i)
}

func (i *InterceptedMessage) Unmarshal(b []byte) error {
	return json.Unmarshal(b, i)
}

type timeout struct {
	Type     string `json:"type"`
	Duration int    `json:"duration"`
	Peer     PeerID `json:"peer"`
}

type request struct {
	Type    string             `json:"type"`
	Peer    *Peer              `json:"peer,omitempty"`
	Message InterceptedMessage `json:"message,omitempty"`
	Timeout *timeout           `json:"timeout,omitempty"`
}

type response struct {
	Status string `json:"status"`
	Err    string `json:"error"`
}

func Unmarshal(msg string) (*request, *types.Error) {
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

type DirectiveMessage struct {
	Action string `json:"action"`
}

var (
	RestartDirective = DirectiveMessage{Action: "RESTART"}
	IsReadyDirective = DirectiveMessage{Action: "ISREADY"}
)
