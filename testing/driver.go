package testing

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/ds-test-framework/scheduler/algos/common"
	"github.com/ds-test-framework/scheduler/log"
	transport "github.com/ds-test-framework/scheduler/transports/http"
	"github.com/ds-test-framework/scheduler/types"
)

type testDriver struct {
	ctx          *types.Context
	messages     chan types.ContextEvent
	stateUpdates chan types.ContextEvent
	logs         chan types.ContextEvent
	msgStore     *common.MsgStore
	dispatchCh   chan types.ReplicaID
	stopCh       chan bool
	totalPeers   int

	curTestCase TestCase
	mtx         *sync.Mutex

	logger *log.Logger
}

func newTestDriver(ctx *types.Context) *testDriver {
	config := ctx.Config("root")
	config.SetDefault("num_replicas", 4)

	d := &testDriver{
		ctx:          ctx,
		messages:     ctx.Subscribe(types.InterceptedMessage),
		stateUpdates: ctx.Subscribe(types.StateMessage),
		logs:         ctx.Subscribe(types.LogMessage),
		dispatchCh:   make(chan types.ReplicaID, 10),
		stopCh:       make(chan bool),
		totalPeers:   config.GetInt("num_replicas"),

		curTestCase: nil,
		mtx:         new(sync.Mutex),
		logger:      ctx.Logger.With(map[string]interface{}{"service": "test-driver"}),
	}
	d.msgStore = common.NewMsgStore(d.dispatchCh)
	return d
}

func (d *testDriver) Run() {
	d.logger.Debug("Starting test driver")
	go d.pollAPI()
	go d.pollDispatch()
}

func (d *testDriver) Stop() {
	d.logger.Debug("Stopping test driver")
	close(d.stopCh)
}

func (d *testDriver) StartRun(t TestCase) {
	d.logger.With(map[string]interface{}{"testcase": t.Name()}).Debug("Setting up driver for test case")

	d.msgStore.Reset()
	d.ctx.Replicas.ResetReady()
	d.mtx.Lock()
	d.curTestCase = t
	d.mtx.Unlock()

	d.restartReplicas()
}

func (d *testDriver) StopRun() {
	d.msgStore.Reset()
	d.mtx.Lock()
	t := d.curTestCase
	d.curTestCase = nil
	d.mtx.Unlock()

	d.logger.With(map[string]interface{}{"testcase": t.Name()}).Debug("Tearing down driver for test case")
}

func (m *testDriver) Ready() bool {
	for {
		select {
		case <-m.stopCh:
			return false
		default:
		}

		if m.ctx.Replicas.NumReady() == m.totalPeers {
			return true
		}
	}
}

func (d *testDriver) restartReplicas() {
	var wg sync.WaitGroup
	for _, p := range d.ctx.Replicas.Iter() {
		wg.Add(1)
		go func(peer *types.Replica) {
			d.sendPeerDirective(peer, &restartDirective)
			wg.Done()
		}(p)
	}
	wg.Wait()
}

func (d *testDriver) pollAPI() {
	for {
		select {
		case event := <-d.messages:
			if event.Type == types.InterceptedMessage {
				go d.handleMessage(event)
			}
		case event := <-d.stateUpdates:
			if event.Type == types.StateMessage {
				go d.handleStateUpdate(event)
			}
		case event := <-d.logs:
			if event.Type == types.LogMessage {
				go d.handleLogMessage(event)
			}
		case <-d.stopCh:
			return
		}
	}
}

func (d *testDriver) handleMessage(event types.ContextEvent) {
	msg, ok := event.Data.(*types.MessageWrapper)
	if !ok {
		return
	}
	d.msgStore.Add(msg.Msg)
	d.mtx.Lock()
	testCase := d.curTestCase
	d.mtx.Unlock()
	if testCase == nil {
		return
	}

	pass, more := testCase.HandleMessage(msg.Msg.Clone())
	if pass {
		d.msgStore.Mark(msg.Msg.ID)
	}
	for _, msgN := range more {
		if msgN.ID == "" {
			msgN.ID = strconv.Itoa(d.ctx.IDGen.Next())
		}
		if !d.msgStore.Exists(msgN.ID) {
			d.msgStore.Add(msgN)
		}
		d.msgStore.Mark(msgN.ID)
	}
}

func (d *testDriver) handleStateUpdate(event types.ContextEvent) {
	state, ok := event.Data.(*types.StateUpdate)
	if !ok {
		return
	}
	d.mtx.Lock()
	testCase := d.curTestCase
	d.mtx.Unlock()
	if testCase == nil {
		return
	}

	testCase.HandleStateUpdate(state)
}

func (d *testDriver) handleLogMessage(event types.ContextEvent) {
	log, ok := event.Data.(*types.ReplicaLog)
	if !ok {
		return
	}
	d.mtx.Lock()
	testCase := d.curTestCase
	d.mtx.Unlock()
	if testCase == nil {
		return
	}

	testCase.HandleLogMessage(log)
}

func (d *testDriver) dispatch(replicaID types.ReplicaID) {
	replica, exists := d.ctx.Replicas.GetReplica(replicaID)
	if !exists {
		return
	}

	msg, err := d.msgStore.FetchOne(replicaID)
	if err != nil {
		return
	}

	ok := d.msgStore.Exists(msg.ID)
	if !ok {
		return
	}

	if msg.Timeout {
		go d.sendPeerTimeout(replica, msg.Type)
	} else {
		go d.sendPeerMsg(replica, msg)
	}

}

func (d *testDriver) pollDispatch() {
	for {
		select {
		case replicaID := <-d.dispatchCh:
			go d.dispatch(replicaID)
		case <-d.stopCh:
			return
		}
	}
}

// DirectiveMessage represents an action that a given replica should perform
type directiveMessage struct {
	Action string `json:"action"`
}

var (
	restartDirective = directiveMessage{Action: "RESTART"}
	isReadyDirective = directiveMessage{Action: "ISREADY"}
)

func (m *testDriver) sendPeerMsg(peer *types.Replica, msg *types.Message) {
	b, err := json.Marshal(msg)
	if err != nil {
		return
	}
	// logger.Debug(fmt.Sprintf("Sending peer message: %s", string(b)))
	transport.SendMsg(http.MethodPost, peer.Addr+"/message", string(b), transport.JsonRequest())
}

func (m *testDriver) sendPeerDirective(peer *types.Replica, d *directiveMessage) (string, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return "", err
	}
	return transport.SendMsg(http.MethodPost, peer.Addr+"/directive", string(b), transport.JsonRequest())
}

func (m *testDriver) sendPeerTimeout(peer *types.Replica, t string) {
	timeout := &types.Timeout{
		Type: t,
	}
	b, err := json.Marshal(timeout)
	if err != nil {
		return
	}
	transport.SendMsg(http.MethodPost, peer.Addr+"/timeout", string(b), transport.JsonRequest())
}
