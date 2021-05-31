package tendermint

import (
	"github.com/ds-test-framework/scheduler/types"
)

type WorkloadInjector struct {
	ctx *types.Context
}

func NewWorkloadInjector(ctx *types.Context) *WorkloadInjector {
	return &WorkloadInjector{
		ctx: ctx,
	}
}

func (w *WorkloadInjector) InjectWorkLoad() *types.Error {
	// resp, err := transport.SendMsg(http.MethodGet, "localhost:26657/broadcast_tx_commit?tx=\"one=two\"", "")
	// if err != nil {
	// 	return err
	// }
	// logger.Debug(fmt.Sprintf("Tendermint workload resp: %s", resp))
	return nil
}
