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
	ErrDestUnknown   = errors.New("destination unknown")
	ErrFailedMarshal = errors.New("failed to marshal data")
)

type Dispatcher struct {
	Replicas *types.ReplicaStore
}

func NewDispatcher(ctx *context.RootContext) *Dispatcher {
	return &Dispatcher{
		Replicas: ctx.Replicas,
	}
}

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

func (d *Dispatcher) StopReplica(replica types.ReplicaID) error {
	return errors.New("not implemented")
}

func (d *Dispatcher) StartReplica(replica types.ReplicaID) error {
	return errors.New("not implemented")
}

func (d *Dispatcher) RestartReplica(replica types.ReplicaID) error {
	return errors.New("not implemented")
}
