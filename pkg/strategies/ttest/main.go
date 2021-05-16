package ttest

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ds-test-framework/scheduler/pkg/log"
	"github.com/ds-test-framework/scheduler/pkg/types"
	"github.com/gogo/protobuf/proto"
	tmsg "github.com/tendermint/tendermint/proto/tendermint/consensus"
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
		filters:      []Filter{NewBlockNAllow("BlockPart", 5, 10)},

		ctx:    ctx,
		inChan: ctx.Subscribe(types.ScheduledMessage),
		logger: ctx.Logger.With(map[string]string{
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
	n.logger.Debug(fmt.Sprintf("Message types received: %#v", n.messageTypes))
	n.mLock.Unlock()
	close(n.stopCh)
}

func (n *TTestScheduler) poll() {
	for {
		select {
		case event := <-n.inChan:
			n.logger.With(map[string]string{
				"event_type":        event.Type.String(),
				"scheduled_message": fmt.Sprintf("%#v", event.Message.Msg),
				"run":               fmt.Sprintf("%d", event.Message.Run),
			}).Debug("Received message")
			m := event.Message
			cMsg, err := unmarshal(m.Msg.Msg)
			ok := true
			if err == nil {
				// n.logger.Debug(fmt.Sprintf("Message on channel id: %d, %s", cMsg.ChannelID, cMsg.MsgB))
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
					n.ctx.MarkMessage(m)
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

	// log.Debug(fmt.Sprintf("Received message from: %s, with contents: %s", cMsg.From, cMsg.Msg.String()))
	return &cMsg, err
}
