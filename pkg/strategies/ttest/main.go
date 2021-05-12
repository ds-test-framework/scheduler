package ttest

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ds-test-framework/scheduler/pkg/logger"
	"github.com/ds-test-framework/scheduler/pkg/types"
	"github.com/gogo/protobuf/proto"
	tmsg "github.com/tendermint/tendermint/proto/tendermint/consensus"
)

// NopScheduler does nothing. Just returns the incoming message in the outgoing channel
type TTestScheduler struct {
	inChan  chan *types.MessageWrapper
	outChan chan *types.MessageWrapper
	stopCh  chan bool

	messageTypes map[string]int
	mLock        *sync.Mutex

	filters []Filter
}

// NewTTestScheduler returns a new TTestScheduler
func NewTTestScheduler() *TTestScheduler {
	return &TTestScheduler{
		stopCh:       make(chan bool, 1),
		messageTypes: make(map[string]int),
		mLock:        new(sync.Mutex),
		filters:      []Filter{NewBlockNAllow("BlockPart", 5, 10)},
	}
}

// Reset implements StrategyEngine
func (n *TTestScheduler) Reset() {
}

// Run implements StrategyEngine
func (n *TTestScheduler) Run() *types.Error {
	logger.Debug("Starting TTestScheduler")
	go n.poll()
	return nil
}

// Stop implements StrategyEngine
func (n *TTestScheduler) Stop() {
	logger.Debug("Stopping TTestScheduler")
	n.mLock.Lock()
	logger.Debug(fmt.Sprintf("Message types received: %#v", n.messageTypes))
	n.mLock.Unlock()
	close(n.stopCh)
}

// SetChannels implements StrategyEngine
func (n *TTestScheduler) SetChannels(inChan chan *types.MessageWrapper, outChan chan *types.MessageWrapper) {
	n.inChan = inChan
	n.outChan = outChan
}

func (n *TTestScheduler) poll() {
	for {
		select {
		case m := <-n.inChan:
			cMsg, err := unmarshal(m.Msg.Msg)
			ok := true
			if err == nil {
				// logger.Debug(fmt.Sprintf("Message on channel id: %d, %s", cMsg.ChannelID, cMsg.MsgB))
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
				logger.Debug("Error unmarshalling: " + err.Error())
			}
			if ok {
				go func(m *types.MessageWrapper) {
					n.outChan <- m
				}(m)
			}
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

func unmarshal(m []byte) (*ControllerMsgEnvelop, error) {
	var cMsg ControllerMsgEnvelop
	err := json.Unmarshal(m, &cMsg)
	if err != nil {
		return &cMsg, err
	}

	msg := proto.Clone(new(tmsg.Message))
	msg.Reset()

	if err := proto.Unmarshal(cMsg.MsgB, msg); err != nil {
		logger.Debug("Error unmarshalling")
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
		cMsg.Type = v.Vote.Type.String()
	case *tmsg.Message_HasVote:
		cMsg.Type = "HasVote"
	case *tmsg.Message_VoteSetMaj23:
		cMsg.Type = "VoteSetMaj23"
	case *tmsg.Message_VoteSetBits:
		cMsg.Type = "VoteSetBits"
	default:
		cMsg.Type = "None"
	}

	logger.Debug(fmt.Sprintf("Received message from: %s, with contents: %s", cMsg.From, cMsg.Msg.String()))
	return &cMsg, err
}
