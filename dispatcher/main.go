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

	// Directive action to start the replica
	startAction = &directiveMessage{
		Action: "START",
	}

	// Directive action to stop the replica
	stopAction = &directiveMessage{
		Action: "STOP",
	}

	// Directive action to restart the replica
	restartAction = &directiveMessage{
		Action: "RESTART",
	}
)

type directiveMessage struct {
	Action string `json:"action"`
}

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
	replicaS, ok := d.Replicas.Get(replica)
	if !ok {
		return ErrDestUnknown
	}
	return d.sendDirective(stopAction, replicaS.Addr)
}

// StartReplica should be called to direct the replica to start running
func (d *Dispatcher) StartReplica(replica types.ReplicaID) error {
	replicaS, ok := d.Replicas.Get(replica)
	if !ok {
		return ErrDestUnknown
	}
	return d.sendDirective(startAction, replicaS.Addr)
}

// RestartReplica should be called to direct the replica to restart
func (d *Dispatcher) RestartReplica(replica types.ReplicaID) error {
	replicaS, ok := d.Replicas.Get(replica)
	if !ok {
		return ErrDestUnknown
	}
	return d.sendDirective(restartAction, replicaS.Addr)
}

func (d *Dispatcher) RestartAll() error {
	errCh := make(chan error, d.Replicas.Cap())
	for _, r := range d.Replicas.Iter() {
		go func(errCh chan error, replicaAddr string) {
			errCh <- d.sendDirective(restartAction, replicaAddr)
		}(errCh, r.Addr)
	}
	for i := 0; i < d.Replicas.Cap(); i++ {
		err := <-errCh
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Dispatcher) sendDirective(directive *directiveMessage, addr string) error {
	bytes, err := json.Marshal(directive)
	if err != nil {
		return ErrFailedMarshal
	}
	_, err = util.SendMsg(http.MethodPost, addr+"/directive", string(bytes), util.JsonRequest())
	if err != nil {
		return err
	}
	return nil
}
