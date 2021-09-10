package dispatcher

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ds-test-framework/scheduler/context"
	"github.com/ds-test-framework/scheduler/types"
	"github.com/ds-test-framework/scheduler/util"
)

var (
	// ErrDestUnknown is returned when a message is asked to be dispatched to an unknown replicas
	ErrDestUnknown = errors.New("destination unknown")
	// ErrFailedMarshal is returned when the message could not be marshalled
	ErrFailedMarshal = errors.New("failed to marshal data")
)

// Dispatcher should be used to send messages to the replicas and handles the marshalling of messages
type Dispatcher struct {
	// Replicas an instance of the replica store
	Replicas *types.ReplicaStore
}

// NewDispatcher instantiates a new instance of Dispatcher
func NewDispatcher(ctx *context.RootContext) *Dispatcher {
	return &Dispatcher{
		Replicas: ctx.Replicas,
	}
}

// DispatchMessage should be called to send an _intercepted_ message to a replica
func (d *Dispatcher) DispatchMessage(msg *types.Message) error {
	replica, ok := d.Replicas.Get(msg.To)
	if !ok {
		return ErrDestUnknown
	}
	bytes, err := json.Marshal(msg)
	if err != nil {
		return ErrFailedMarshal
	}
	_, err = util.SendMsg(http.MethodPost, replica.Addr+"/message", string(bytes), util.JsonRequest())
	if err != nil {
		return err
	}
	return nil
}

// StopReplica should be called to direct the replica to stop running
func (d *Dispatcher) StopReplica(replica types.ReplicaID) error {
	return errors.New("not implemented")
}

// StartReplica should be called to direct the replica to start running
func (d *Dispatcher) StartReplica(replica types.ReplicaID) error {
	return errors.New("not implemented")
}

// RestartReplica should be called to direct the replica to restart
func (d *Dispatcher) RestartReplica(replica types.ReplicaID) error {
	return errors.New("not implemented")
}
