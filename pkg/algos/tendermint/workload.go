package tendermint

import (
	"github.com/ds-test-framework/scheduler/pkg/algos/common"
	"github.com/ds-test-framework/scheduler/pkg/types"
)

type WorkloadInjector struct {
	peers *common.PeerStore
}

func NewWorkloadInjector() *WorkloadInjector {
	return &WorkloadInjector{}
}

func (w *WorkloadInjector) SetPeerStore(p *common.PeerStore) {
	w.peers = p
}

func (w *WorkloadInjector) InjectWorkLoad() *types.Error {
	// resp, err := transport.SendMsg(http.MethodGet, "localhost:26657/broadcast_tx_commit?tx=\"one=two\"", "")
	// if err != nil {
	// 	return err
	// }
	// logger.Debug(fmt.Sprintf("Tendermint workload resp: %s", resp))
	return nil
}
