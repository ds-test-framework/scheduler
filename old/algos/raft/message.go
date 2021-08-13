package raft

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ds-test-framework/scheduler/types"
)

const (
	ErrUnmarshal = "UNMARSHAL_FAILED"
	ErrMashsal   = "MARSHAL_FAILED"
)

type RaftMessage struct {
	From int              `json:"from"`
	To   int              `json:"to"`
	M    TransportMessage `json:"m"`
	ID   int              `json:"id"`
}

type TransportMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type TimeoutMessage struct {
	Time int64  `json:"time"`
	T    string `json:"type"`
	Term int    `json:"term"`
}

func UnmarshalMessage(message string) (*types.Message, *types.Error) {
	var msg RaftMessage
	err := json.Unmarshal([]byte(message), &msg)
	if err != nil {
		return nil, types.NewError(
			ErrUnmarshal,
			"Could not unmarshall message",
		)
	}
	timeout := false
	weight := 0
	if msg.M.Type == "Timeout" {
		timeout = true
		var timeoutMsg TimeoutMessage
		err := json.Unmarshal([]byte(msg.M.Message), &timeoutMsg)
		if err == nil {
			weight = int(timeoutMsg.Time)
		}
	}

	return types.NewMessage(
		msg.M.Type,
		strconv.FormatInt(int64(msg.ID), 10),
		types.ReplicaID(strconv.Itoa(msg.From)),
		types.ReplicaID(strconv.Itoa(msg.To)),
		weight,
		timeout,
		[]byte(msg.M.Message),
		true,
	), nil

}

type ControlType string

const (
	Ready          ControlType = "Ready"
	Reset          ControlType = "Reset"
	ReadyToInject  ControlType = "ReadyToInject"
	InjectWorkload ControlType = "InjectWorkload"
)

type ControlMessage struct {
	Type string `json:"type"`
}

func NewControlMessage(cType ControlType) *ControlMessage {
	return &ControlMessage{
		Type: string(cType),
	}
}

func (m *ControlMessage) Marshall() (string, *types.Error) {
	b, err := json.Marshal(m)
	if err != nil {
		return string(b), types.NewError(
			ErrMashsal,
			"Could not marshall control message",
		)
	}
	return string(b), nil
}

func UnmarshalControlMessage(m string) (*ControlMessage, *types.Error) {
	var cMsg ControlMessage
	err := json.Unmarshal([]byte(m), &cMsg)
	if err != nil {
		return nil, types.NewError(
			ErrUnmarshal,
			fmt.Sprintf("Failed to unmarshall control message: %s", err.Error()),
		)
	}
	return &cMsg, nil
}

func IsControlMessage(m string) bool {
	var cMsg ControlMessage
	err := json.Unmarshal([]byte(m), &cMsg)
	return err != nil
}
