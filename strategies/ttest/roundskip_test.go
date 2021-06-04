package ttest

import (
	"testing"

	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
	tmsg "github.com/tendermint/tendermint/proto/tendermint/consensus"
)

var (
	replicas = []types.ReplicaID{
		types.ReplicaID("one"),
		types.ReplicaID("two"),
		types.ReplicaID("three"),
		types.ReplicaID("four"),
	}

	ctx = &types.Context{
		Replicas: types.NewReplicaStore(4),
		Logger:   log.DummyLogger(),
	}

	msgs = map[string]*tmsg.Message{
		"b20": &tmsg.Message{
			Sum: &tmsg.Message_BlockPart{
				BlockPart: &tmsg.BlockPart{
					Height: 2,
					Round:  0,
				},
			},
		},
		"b10": &tmsg.Message{
			Sum: &tmsg.Message_BlockPart{
				BlockPart: &tmsg.BlockPart{
					Height: 1,
					Round:  0,
				},
			},
		},
		"b11": &tmsg.Message{
			Sum: &tmsg.Message_BlockPart{
				BlockPart: &tmsg.BlockPart{
					Height: 1,
					Round:  1,
				},
			},
		},
		"b12": &tmsg.Message{
			Sum: &tmsg.Message_BlockPart{
				BlockPart: &tmsg.BlockPart{
					Height: 1,
					Round:  2,
				},
			},
		},
		"rs10": &tmsg.Message{
			Sum: &tmsg.Message_NewRoundStep{
				NewRoundStep: &tmsg.NewRoundStep{
					Height: 1,
					Round:  0,
				},
			},
		},
		"rs11": &tmsg.Message{
			Sum: &tmsg.Message_NewRoundStep{
				NewRoundStep: &tmsg.NewRoundStep{
					Height: 1,
					Round:  1,
				},
			},
		},
		"rs12": &tmsg.Message{
			Sum: &tmsg.Message_NewRoundStep{
				NewRoundStep: &tmsg.NewRoundStep{
					Height: 1,
					Round:  2,
				},
			},
		},
	}
)

func TestPRSFPeers(t *testing.T) {
	store := newPeerRoundStore(0, 1)
	store.Record(replicas[0], 0)
	store.Record(replicas[1], 0)

	sreplicas := store.Peers()
	if string(sreplicas[0]) != string(replicas[0]) {
		t.Errorf("Invalid replicas recorded. Expected %s, found %s", replicas[0], sreplicas[0])
	}
	if string(sreplicas[1]) != string(replicas[1]) {
		t.Errorf("Invalid replicas recorded. Expected %s, found %s", replicas[1], sreplicas[1])
	}
}

func TestPRSFPeersCount(t *testing.T) {
	store := newPeerRoundStore(0, 1)
	store.Record(replicas[0], 0)
	store.Record(replicas[0], 0)

	sreplicas := store.Peers()
	if len(sreplicas) != 1 {
		t.Errorf("More than intended replica's recorded: %d", len(sreplicas))
	}
}

func TestPRSRightReplicas(t *testing.T) {
	store := newPeerRoundStore(0, 1)
	store.Record(replicas[0], 0)
	store.Record(replicas[1], 1)
	store.Record(replicas[2], 0)

	sreplicas := store.Peers()
	if len(sreplicas) != 2 {
		t.Errorf("Too many replicas recorded: %d", len(sreplicas))
	}
	if string(sreplicas[0]) != string(replicas[0]) {
		t.Errorf("Invalid replicas recorded. Expected %s, found %s", replicas[0], sreplicas[0])
	}
	if string(sreplicas[1]) != string(replicas[2]) {
		t.Errorf("Invalid replicas recorded. Expected %s, found %s", replicas[1], sreplicas[1])
	}
}

func TestPRSNonReplica(t *testing.T) {
	store := newPeerRoundStore(0, 1)
	store.Record(replicas[0], 0)
	store.Record(replicas[1], 0)
	store.Record(replicas[2], 0)

	sreplicas := store.Peers()
	if len(sreplicas) != 2 {
		t.Errorf("Too many replicas recorded: %d", len(sreplicas))
	}
}

func TestPRSRounSkip(t *testing.T) {
	store := newPeerRoundStore(0, 1)
	store.Record(replicas[0], 0)
	store.Record(replicas[1], 0)
	store.Record(replicas[1], 1)

	if store.skips != 1 {
		t.Errorf("Round skip not recorded")
	}
}

func TestPRSRounSkipSame(t *testing.T) {
	store := newPeerRoundStore(0, 1)
	store.Record(replicas[0], 0)
	store.Record(replicas[1], 0)
	store.Record(replicas[1], 1)
	store.Record(replicas[1], 2)

	if store.skips != 1 {
		t.Errorf("Round skip incorrectly recorded")
	}
}

func TestPRSRoundSkipFlag(t *testing.T) {
	store := newPeerRoundStore(0, 1)
	store.Record(replicas[0], 0)
	store.Record(replicas[1], 0)
	store.Record(replicas[1], 1)
	store.Record(replicas[0], 1)

	if !store.Skipped() {
		t.Errorf("Flag not set when enough replicas have skipped")
	}
}

func TestPRSRoundSkipFlagHigherRound(t *testing.T) {
	store := newPeerRoundStore(1, 1)
	store.Record(replicas[0], 0)
	store.Record(replicas[1], 0)
	store.Record(replicas[1], 1)
	store.Record(replicas[0], 1)

	if store.Skipped() {
		t.Errorf("Flag set when not enough replicas have skipped")
	}
}

func TestPRSRoundSkipFlagNot(t *testing.T) {
	store := newPeerRoundStore(0, 1)
	store.Record(replicas[0], 0)
	store.Record(replicas[1], 0)
	store.Record(replicas[1], 1)

	if store.Skipped() {
		t.Errorf("Flag set when not enough replicas have skipped")
	}
}

func TestRSFHeightNot(t *testing.T) {
	filter := NewRoundSkipFilter(ctx, 1, 1)
	res := filter.Test(&ControllerMsgEnvelop{
		From: replicas[0],
		To:   replicas[1],
		Msg:  msgs["b20"],
		Type: "BlockPart",
	})
	if !res {
		t.Errorf("Higher height message blocked")
	}
}

// func TestRSFHeight(t *testing.T) {
// 	filter := NewRoundSkipFilter(ctx, 1, 1)
// 	res := filter.recordNCheck(replicas[0], 1, 0)
// 	if !res {
// 		t.Errorf("Message not blocked")
// 	}
// }

// func TestRSFPerfect(t *testing.T) {
// 	filter := NewRoundSkipFilter(ctx, 1, 1)
// 	filter.recordNCheck(replicas[0], 1, 0)
// 	filter.recordNCheck(replicas[1], 1, 0)
// 	filter.recordNCheck(replicas[0], 1, 1)
// 	filter.recordNCheck(replicas[1], 1, 1)

// 	if !filter.flag {
// 		t.Error("Round skip not recorded")
// 	}

// 	filter = NewRoundSkipFilter(ctx, 1, 2)
// 	filter.recordNCheck(replicas[0], 1, 0)
// 	filter.recordNCheck(replicas[1], 1, 0)
// 	filter.recordNCheck(replicas[0], 1, 1)
// 	filter.recordNCheck(replicas[1], 1, 1)

// 	if filter.flag {
// 		t.Error("Round skip incorrectly recoreded")
// 	}

// 	filter.recordNCheck(replicas[0], 1, 2)
// 	filter.recordNCheck(replicas[1], 1, 2)
// 	if !filter.flag {
// 		t.Error("Round skip not recoreded")
// 	}
// }

// func TestRSFRoundStateUpdate(t *testing.T) {
// 	filter := NewRoundSkipFilter(ctx, 1, 2)
// 	filter.recordNCheck(replicas[0], 1, 0)
// 	filter.recordNCheck(replicas[1], 1, 0)
// 	filter.recordNCheck(replicas[0], 1, 1)
// 	filter.recordNCheck(replicas[1], 1, 1)

// 	if filter.curRound != 1 {
// 		t.Error("Round skip not recorded")
// 	}
// }

// func TestRSFCurRound(t *testing.T) {
// 	filter := NewRoundSkipFilter(ctx, 1, 2)
// 	filter.recordNCheck(replicas[0], 1, 0)
// 	filter.recordNCheck(replicas[1], 1, 0)
// 	filter.recordNCheck(replicas[0], 1, 1)
// 	filter.recordNCheck(replicas[1], 1, 1)
// 	filter.recordNCheck(replicas[0], 1, 2)
// 	filter.recordNCheck(replicas[1], 1, 2)
// 	if filter.curRound != 2 {
// 		t.Error("Round skip not recorded")
// 	}
// }

// func TestRSFBlockingHigherRound(t *testing.T) {
// 	filter := NewRoundSkipFilter(ctx, 1, 2)
// 	filter.recordNCheck(replicas[0], 1, 0)
// 	filter.recordNCheck(replicas[1], 1, 0)
// 	filter.recordNCheck(replicas[0], 1, 1)
// 	filter.recordNCheck(replicas[1], 1, 1)
// 	block := filter.recordNCheck(replicas[0], 1, 2)

// 	if !block {
// 		t.Error("Replica round state not recorded")
// 	}
// }

// func TestRSFRoundUpdate(t *testing.T) {
// 	filter := NewRoundSkipFilter(ctx, 1, 2)
// 	filter.recordNCheck(replicas[0], 1, 0)
// 	filter.recordNCheck(replicas[1], 1, 0)
// 	filter.recordNCheck(replicas[0], 1, 1)
// 	filter.recordNCheck(replicas[1], 1, 1)
// 	filter.recordNCheck(replicas[0], 1, 2)
// 	if len(filter.roundStore) < 2 {
// 		t.Error("New peerRoundStore not created when round updates")
// 	}
// }

// func TestRSFFlagNot(t *testing.T) {
// 	filter := NewRoundSkipFilter(ctx, 1, 1)
// 	filter.recordNCheck(replicas[0], 1, 0)
// 	filter.recordNCheck(replicas[1], 1, 0)
// 	filter.recordNCheck(replicas[0], 1, 1)

// 	if filter.flag {
// 		t.Error("Round skip incorrectly recorded")
// 	}
// }

func TestRSFMessages(t *testing.T) {
	filter := NewRoundSkipFilter(ctx, 1, 1)

	ok := filter.Test(&ControllerMsgEnvelop{
		From: replicas[0],
		To:   replicas[1],
		Msg:  msgs["rs10"],
		Type: "NewRoundStep",
	})
	if !ok {
		t.Errorf("RoundStep message blocked")
	}

	filter.Test(&ControllerMsgEnvelop{
		From: replicas[1],
		To:   replicas[0],
		Msg:  msgs["rs10"],
		Type: "NewRoundStep",
	})

	ok = filter.Test(&ControllerMsgEnvelop{
		From: replicas[2],
		To:   replicas[0],
		Msg:  msgs["b10"],
		Type: "BlockPart",
	})
	if ok {
		t.Error("BlockPart message not blocked")
	}
}
