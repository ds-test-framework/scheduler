package tendermint

import (
	"net/http"

	"github.com/ds-test-framework/scheduler/pkg/algos/common"
	transport "github.com/ds-test-framework/scheduler/pkg/transports/http"
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
	var peer *common.Replica
	for _, p := range w.peers.Iter() {
		peer = p
		break
	}
	_, err := transport.SendMsg(http.MethodGet, peer.Addr+`/broadcast_tx_commit?tx="name=srinidhi"`, "")
	return err
}
