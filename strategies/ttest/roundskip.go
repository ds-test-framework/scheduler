package ttest

import (
	"sync"

	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
)

type peerRoundStore struct {
	round        int
	flag         bool
	skips        int
	faults       int
	fPeers       map[types.ReplicaID]int
	peersSkipped map[types.ReplicaID]bool

	mtx *sync.Mutex
}

func newPeerRoundStore(round, faults int) *peerRoundStore {
	return &peerRoundStore{
		round:        round,
		flag:         false,
		skips:        0,
		faults:       faults,
		fPeers:       make(map[types.ReplicaID]int),
		peersSkipped: make(map[types.ReplicaID]bool),
		mtx:          new(sync.Mutex),
	}
}

func (p *peerRoundStore) AddPeer(peer types.ReplicaID) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	if len(p.fPeers) == p.faults+1 {
		return
	}
	p.fPeers[peer] = p.round
}

func (p *peerRoundStore) Peers() []types.ReplicaID {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	replicas := make([]types.ReplicaID, p.faults+1)
	i := 0
	for peer := range p.fPeers {
		replicas[i] = peer
		i += 1
	}
	return replicas
}

func (p *peerRoundStore) Skipped() bool {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	return p.flag
}

func (p *peerRoundStore) Record(peer types.ReplicaID, round int) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	curRound, ok := p.fPeers[peer]
	if !ok {
		return
	}
	if round > curRound {
		p.fPeers[peer] = round
		if !p.peersSkipped[peer] {
			p.peersSkipped[peer] = true
			p.skips = p.skips + 1
			if p.skips == p.faults+1 {
				p.flag = true
			}
		}
	}
}

type RoundSkipFilter struct {
	mtx *sync.Mutex

	fPeer      map[types.ReplicaID]bool
	fPeerCount int

	faults       int
	roundStore   map[int]*peerRoundStore
	curRound     int
	flag         bool
	height       int
	roundsToSkip int

	logger *log.Logger
}

func NewRoundSkipFilter(ctx *types.Context, height int, roundsToSkip int) *RoundSkipFilter {
	faults := int((ctx.Replicas.Size() - 1) / 3)

	return &RoundSkipFilter{
		mtx:          new(sync.Mutex),
		fPeer:        make(map[types.ReplicaID]bool),
		fPeerCount:   0,
		faults:       faults,
		roundStore:   make(map[int]*peerRoundStore),
		roundsToSkip: roundsToSkip,
		curRound:     0,
		flag:         false,
		height:       height,

		logger: ctx.Logger.With(map[string]interface{}{
			"service": "roundskip-filter",
		}),
	}
}

func (e *RoundSkipFilter) extractHR(msg *ControllerMsgEnvelop) (int, int) {
	switch msg.Type {
	case "NewRoundStep":
		hrs := msg.Msg.GetNewRoundStep()
		return int(hrs.Height), int(hrs.Round)
	case "Proposal":
		prop := msg.Msg.GetProposal()
		return int(prop.Proposal.Height), int(prop.Proposal.Round)
	case "Prevote":
		vote := msg.Msg.GetVote()
		return int(vote.Vote.Height), int(vote.Vote.Round)
	case "Precommit":
		vote := msg.Msg.GetVote()
		return int(vote.Vote.Height), int(vote.Vote.Round)
	case "Vote":
		vote := msg.Msg.GetVote()
		return int(vote.Vote.Height), int(vote.Vote.Round)
	case "NewValidBlock":
		block := msg.Msg.GetNewValidBlock()
		return int(block.Height), int(block.Round)
	case "ProposalPol":
		pPol := msg.Msg.GetProposalPol()
		return int(pPol.Height), -1
	case "VoteSetMaj23":
		vote := msg.Msg.GetVoteSetMaj23()
		return int(vote.Height), int(vote.Round)
	case "VoteSetBits":
		vote := msg.Msg.GetVoteSetBits()
		return int(vote.Height), int(vote.Round)
	case "BlockPart":
		blockPart := msg.Msg.GetBlockPart()
		return int(blockPart.Height), int(blockPart.Round)

	}
	return -1, -1
}

func (e *RoundSkipFilter) newStore(round int) {
	store := newPeerRoundStore(round, e.faults)
	for peer := range e.fPeer {
		store.AddPeer(peer)
	}
	e.roundStore[round] = store
}

func (e *RoundSkipFilter) record(peer types.ReplicaID, round int) {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	if e.fPeerCount != e.faults+1 {
		_, ok := e.fPeer[peer]
		if !ok {
			e.fPeer[peer] = true
			for _, store := range e.roundStore {
				store.AddPeer(peer)
			}
		}
	}

	_, ok := e.roundStore[e.curRound]
	if !ok {
		e.newStore(e.curRound)
	}

	if round > e.curRound {
		_, ok = e.roundStore[round]
		if !ok {
			e.newStore(round)
		}
		roundStore := e.roundStore[round]
		roundStore.Record(peer, round)
	}

	curRoundStore := e.roundStore[e.curRound]
	curRoundStore.Record(peer, round)

	if curRoundStore.Skipped() {
		e.curRound = e.curRound + 1
		if e.curRound == e.roundsToSkip {
			e.flag = true
		}
	}

}

func (e *RoundSkipFilter) checkPeer(peer types.ReplicaID) bool {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	_, ok := e.fPeer[peer]
	return ok
}

func (e *RoundSkipFilter) Test(msg *ControllerMsgEnvelop) bool {
	if msg.Type == "None" {
		return true
	}

	height, round := e.extractHR(msg)
	if height == -1 || round == -1 {
		return true
	}

	e.mtx.Lock()
	curHeight := e.height
	flag := e.flag
	e.mtx.Unlock()
	if flag || height != curHeight {
		return true
	}
	// Record data about peer. (Round number from any vote message or NewRoundStep)
	// If its not something we should check for then skip
	e.record(msg.From, round)

	block := e.checkPeer(msg.To)
	if !block {
		return true
	}

	// BlockPart messages should be blocked and the rest allowed
	return msg.Type != "BlockPart"
}
