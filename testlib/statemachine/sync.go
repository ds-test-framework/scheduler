package statemachine

import (
	"sync"

	"github.com/ds-test-framework/scheduler/testlib"
	"github.com/ds-test-framework/scheduler/types"
)

var _ testlib.Handler = &SyncStateMachineHandler{}

// Round abstracts the stages of a protocol
// In order to force synchronous executions, we cache and send messages only of one particular round
// Round should mandatorily compare with NilRound and return 1 always
type Round interface {
	// -1 for cur < s
	// 0 for cur == s
	// 1 for cur > s
	Compare(s Round) int
	// Empty when the round is inexistant
	// To be used for those messages which do not belong to a particular round
	Empty() bool
	// Key index of the round, to easily compare between different instances of the same round
	Key() string
}

type NilRound struct {
}

func (n *NilRound) Compare(r Round) int {
	switch r.(type) {
	case *NilRound:
		return 0
	}
	return -1
}
func (n *NilRound) Empty() bool {
	return true
}
func (n *NilRound) Key() string {
	return "_nilRound"
}

type MsgToRound func(*types.Message) Round

type msgStore struct {
	Round Round
	Store *types.MessageStore
}

func newMsgStore(round Round) *msgStore {
	return &msgStore{
		Round: round,
		Store: types.NewMessageStore(),
	}
}

type SyncStateMachineHandler struct {
	EventHandlers []EventHandler
	StateMachine  *StateMachine
	extractRound  MsgToRound

	undeliveredMessages map[string]*msgStore
	roundTracker        map[types.ReplicaID]Round
	curRound            Round
	smallest            Round
	lock                *sync.Mutex
}

func NewSyncStateMachineHandler(stateMachine *StateMachine, extractRound MsgToRound, roundZero Round) *SyncStateMachineHandler {
	s := &SyncStateMachineHandler{
		StateMachine:        stateMachine,
		extractRound:        extractRound,
		undeliveredMessages: make(map[string]*msgStore),
		roundTracker:        make(map[types.ReplicaID]Round),
		lock:                new(sync.Mutex),
		curRound:            roundZero,
		smallest:            &NilRound{},
	}
	nilRound := &NilRound{}
	s.undeliveredMessages[nilRound.Key()] = newMsgStore(nilRound)
	return s
}

func (s *SyncStateMachineHandler) countAndUpdateRound(r Round, from types.ReplicaID) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Assumption made here is that replica only moves to higher rounds
	// We cannot receive a message from a replica of round r1 after we have seen a message in round r2 where r1 < r2
	s.roundTracker[from] = r

	smallest := r
	for _, curRound := range s.roundTracker {
		if smallest.Compare(curRound) == -1 {
			smallest = curRound
		}
	}
	if smallest.Compare(s.curRound) == 1 {
		s.curRound = smallest
	}

}

func (s *SyncStateMachineHandler) HandleEvent(e *types.Event, c *testlib.Context) []*types.Message {
	// Group messages to rounds and keep track before delivering messages

	if e.IsMessageSend() {
		messageID, _ := e.MessageID()
		message, ok := c.MessagePool.Get(messageID)
		if ok {
			round := s.extractRound(message)
			if round.Empty() {
				nilRound := &NilRound{}
				s.lock.Lock()
				s.undeliveredMessages[nilRound.Key()].Store.Add(message)
				s.lock.Unlock()
			} else {
				s.lock.Lock()
				_, ok := s.undeliveredMessages[round.Key()]
				if !ok {
					s.undeliveredMessages[round.Key()] = newMsgStore(round)
				}
				s.undeliveredMessages[round.Key()].Store.Add(message)
				s.lock.Unlock()
				s.countAndUpdateRound(round, message.From)
			}
		}
	}

	handled := false
	result := make([]*types.Message, 0)
	ctx := wrapContext(c, s.StateMachine)
	for _, handler := range s.EventHandlers {
		hResponse, ok := handler(e, ctx)
		if ok {
			handled = true
			result = hResponse
			break
		}
	}
	if !handled {
		result, _ = defaultSendHandler(e, ctx)
	}

	for _, message := range result {
		round := s.extractRound(message)
		_, ok := s.undeliveredMessages[round.Key()]
		if !ok {
			s.undeliveredMessages[round.Key()] = newMsgStore(round)
		}
		store := s.undeliveredMessages[round.Key()].Store
		if !store.Exists(message.ID) {
			store.Add(message)
		}
	}

	newState := s.StateMachine.CurState()
	if newState.Label == FailStateLabel {
		c.Abort()
	}

	return s.deliverMessages()
}

func (s *SyncStateMachineHandler) deliverMessages() []*types.Message {
	s.lock.Lock()
	defer s.lock.Unlock()

	curRound := s.curRound
	result := make([]*types.Message, 0)
	for _, messages := range s.undeliveredMessages {
		cmp := curRound.Compare(messages.Round)
		if cmp == 0 || cmp == 1 {
			result = append(result, messages.Store.Iter()...)
			messages.Store.RemoveAll()
		}
	}
	return result
}

func (s *SyncStateMachineHandler) Name() string {
	return "SyncStateMachineHandler"
}

func (s *SyncStateMachineHandler) AddHandler(handler EventHandler) {
	s.EventHandlers = append(s.EventHandlers, handler)
}

func (s *SyncStateMachineHandler) Finalize() {}
