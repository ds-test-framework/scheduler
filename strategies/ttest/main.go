package ttest

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
	"github.com/gogo/protobuf/proto"
	tmsg "github.com/tendermint/tendermint/proto/tendermint/consensus"
	ttypes "github.com/tendermint/tendermint/proto/tendermint/types"
)

// NopScheduler does nothing. Just returns the incoming message in the outgoing channel
type TTestScheduler struct {
	inChan chan types.ContextEvent
	ctx    *types.Context
	stopCh chan bool

	messageTypes map[string]int
	mLock        *sync.Mutex

	logger *log.Logger

	filters []Filter
}

// NewTTestScheduler returns a new TTestScheduler
func NewTTestScheduler(ctx *types.Context) *TTestScheduler {
	return &TTestScheduler{
		stopCh:       make(chan bool, 1),
		messageTypes: make(map[string]int),
		mLock:        new(sync.Mutex),
		filters:      []Filter{NewRoundSkipFilter(ctx)}, // NewBlockNAllow("BlockPart", 5, 10)

		ctx:    ctx,
		inChan: ctx.Subscribe(types.ScheduledMessage),
		logger: ctx.Logger.With(map[string]interface{}{
			"service": "TTestScheduler",
		}),
	}
}

// Reset implements StrategyEngine
func (n *TTestScheduler) Reset() {
}

// Start implements StrategyEngine
func (n *TTestScheduler) Start() *types.Error {
	n.logger.Debug("Starting TTestScheduler")
	go n.poll()
	return nil
}

// Stop implements StrategyEngine
func (n *TTestScheduler) Stop() {
	n.logger.Debug("Stopping TTestScheduler")
	n.mLock.Lock()
	n.logger.Info(fmt.Sprintf("Message types received: %#v", n.messageTypes))
	n.mLock.Unlock()
	close(n.stopCh)
}

func (n *TTestScheduler) handleIncoming(event types.ContextEvent) {
	if event.Type != types.ScheduledMessage {
		return
	}
	msg, ok := event.Data.(*types.MessageWrapper)
	if !ok {
		return
	}
	n.logger.With(map[string]interface{}{
		"event_type": event.Type.String(),
		"msg_id":     msg.Msg.ID,
		"run":        fmt.Sprintf("%d", msg.Run),
	}).Debug("Received message")
	m := msg
	cMsg, err := unmarshal(m.Msg.Msg)
	ok = true
	if err == nil {
		n.logger.Debug(fmt.Sprintf("Received message from: %s, with contents: %s", cMsg.From, cMsg.Msg.String()))
		n.mLock.Lock()
		n.messageTypes[cMsg.Type] = n.messageTypes[cMsg.Type] + 1
		n.mLock.Unlock()

		for _, f := range n.filters {
			if !f.Test(cMsg) {
				ok = false
				break
			}
		}
	} else {
		n.logger.Debug("Error unmarshalling: " + err.Error())
	}
	if ok {
		go func(m *types.MessageWrapper) {
			n.logger.With(map[string]interface{}{
				"to":     string(m.Msg.To),
				"msg_id": m.Msg.ID,
			}).Debug("Dispatching message")
			n.ctx.Publish(types.EnabledMessage, m)
		}(m)
	}
}

func (n *TTestScheduler) poll() {
	for {
		select {
		case event := <-n.inChan:
			go n.handleIncoming(event)
		case <-n.stopCh:
			return
		}
	}
}

type ControllerMsgEnvelop struct {
	ChannelID uint16          `json:"chan_id"`
	MsgB      []byte          `json:"msg"`
	From      types.ReplicaID `json:"from"`
	To        types.ReplicaID `json:"to"`
	Type      string          `json:"-"`
	Msg       *tmsg.Message   `json:"-"`
}

func validChannel(chid uint16) bool {
	// Interpretting messages only of the consensus reactor
	return chid >= 0x20 && chid <= 0x23
}

func unmarshal(m []byte) (*ControllerMsgEnvelop, error) {
	var cMsg ControllerMsgEnvelop
	err := json.Unmarshal(m, &cMsg)
	if err != nil {
		return &cMsg, err
	}

	if !validChannel(cMsg.ChannelID) {
		cMsg.Type = "None"
		return &cMsg, nil
	}

	msg := proto.Clone(new(tmsg.Message))
	msg.Reset()

	if err := proto.Unmarshal(cMsg.MsgB, msg); err != nil {
		// log.Debug("Error unmarshalling")
		cMsg.Type = "None"
		cMsg.Msg = nil
		return &cMsg, nil
	}

	tMsg := msg.(*tmsg.Message)
	cMsg.Msg = tMsg

	switch tMsg.Sum.(type) {
	case *tmsg.Message_NewRoundStep:
		cMsg.Type = "NewRoundStep"
	case *tmsg.Message_NewValidBlock:
		cMsg.Type = "NewValidBlock"
	case *tmsg.Message_Proposal:
		cMsg.Type = "Proposal"
	case *tmsg.Message_ProposalPol:
		cMsg.Type = "ProposalPol"
	case *tmsg.Message_BlockPart:
		cMsg.Type = "BlockPart"
	case *tmsg.Message_Vote:
		v := tMsg.GetVote()
		if v == nil {
			cMsg.Type = "Vote"
			break
		}
		switch v.Vote.Type {
		case ttypes.PrevoteType:
			cMsg.Type = "Prevote"
		case ttypes.PrecommitType:
			cMsg.Type = "Precommit"
		default:
			cMsg.Type = "Vote"
		}
	case *tmsg.Message_HasVote:
		cMsg.Type = "HasVote"
	case *tmsg.Message_VoteSetMaj23:
		cMsg.Type = "VoteSetMaj23"
	case *tmsg.Message_VoteSetBits:
		cMsg.Type = "VoteSetBits"
	default:
		cMsg.Type = "None"
	}

	// log.Debug(fmt.Sprintf("Received message from: %s, with contents: %s", cMsg.From, cMsg.Msg.String()))
	return &cMsg, err
}
