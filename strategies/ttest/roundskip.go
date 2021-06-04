package ttest

import (
	"sync"

	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
)

type RoundSkipFilter struct {
	mtx *sync.Mutex

	fPeer      map[types.ReplicaID]int
	fPeerCount int

	faults       int
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
		fPeer:        make(map[types.ReplicaID]int),
		fPeerCount:   0,
		faults:       faults,
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

func (e *RoundSkipFilter) record(peer types.ReplicaID, round int) {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	if e.fPeerCount != e.faults+1 {
		_, ok := e.fPeer[peer]
		if !ok {
			e.fPeer[peer] = round
		}
	}

	curPeerRound, ok := e.fPeer[peer]
	if !ok {
		return
	}
	if round > curPeerRound {
		e.fPeer[peer] = round
	}

	count := 0
	for _, round := range e.fPeer {
		if round > e.curRound {
			count = count + 1
		}
	}
	if count > e.faults {
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
