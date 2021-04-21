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
	return nil
}
