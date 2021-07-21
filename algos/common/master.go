package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ds-test-framework/scheduler/log"
	transport "github.com/ds-test-framework/scheduler/transports/http"
	"github.com/ds-test-framework/scheduler/types"
)

const (
	ErrPeersNotReady = "ERR_PEERS_NOT_READY"
)

// CommonDriver implements AlgoDriver
// The common driver is agnostic to any specific protocol, it acts as a middleman between the replicas
// and the strategy engine. But also as a central server which pools all the messages that the replicas
// send between each other.
//
// The other role that common driver needs to fulfill is that of controlling the replicas between testing iterations.
// For this purpose the driver sends directives to the replicas to restart/stop or start them. Once the replicas are ready,
// the test harness has to be injected. This is done using the WorkloadInjector interface as it will contain some protocol
// specific logic.
type CommonDriver struct {
	messagesIn       chan types.ContextEvent
	fromEngine       chan types.ContextEvent
	workloadInjector WorkloadInjector

	totalPeers int
	msgStore   *MsgStore

	stopCh   chan bool
	updateCh chan types.ReplicaID

	run     int
	runObj  *types.RunObj
	runLock *sync.Mutex

	ctx    *types.Context
	logger *log.Logger

	ready     bool
	readyLock *sync.Mutex
}

// NewCommonDriver constructs a CommonDriver
// Config options expected
// - type: common
// - algo: for fetching the workload injector of the specific algorithm
// - transport: transport config options
// - num_replicas: number of replicas that are used for testing
func NewCommonDriver(ctx *types.Context, workloadInjector WorkloadInjector) *CommonDriver {
	n := &CommonDriver{
		messagesIn:       ctx.Subscribe(types.InterceptedMessage),
		fromEngine:       ctx.Subscribe(types.EnabledMessage),
		totalPeers:       ctx.Replicas.Size(),
		workloadInjector: workloadInjector,
		stopCh:           make(chan bool),
		updateCh:         make(chan types.ReplicaID, 2),

		run:     0,
		runLock: new(sync.Mutex),
		ctx:     ctx,
		logger: ctx.Logger.With(map[string]interface{}{
			"service": "CommonDriver",
		}),
	}

	n.msgStore = NewMsgStore(n.updateCh)
	return n
}

// Start implements AlgoDriver
func (m *CommonDriver) Start() {
	go m.poll()
}

func (m *CommonDriver) waitForAllPeers() *types.Error {
	timer := time.After(10 * time.Second)

OUTER_LOOP:

	for {
		select {
		case <-timer:
			return types.NewError(
				ErrPeersNotReady,
				"All peers not connected",
			)
		default:
		}
		if m.ctx.Replicas.Count() != m.totalPeers {
			continue
		}
		for _, p := range m.ctx.Replicas.Iter() {
			if !p.Ready {
				continue OUTER_LOOP
			}
		}
		break
	}
	return nil
}

func (m *CommonDriver) Ready() (bool, *types.Error) {
	for {
		select {
		case <-m.stopCh:
			return false, nil
		default:
		}

		if m.ctx.Replicas.NumReady() == m.totalPeers {
			return true, nil
		}
	}
}

func (m *CommonDriver) restartPeers() {
	var wg sync.WaitGroup
	for _, p := range m.ctx.Replicas.Iter() {
		wg.Add(1)
		go func(peer *types.Replica) {
			m.sendPeerDirective(peer, &RestartDirective)
			wg.Done()
		}(p)
	}
	wg.Wait()
}

// StartRun implements AlgoDriver
func (m *CommonDriver) StartRun(run int) (*types.RunObj, *types.Error) {
	err := m.waitForAllPeers()
	if err != nil {
		return nil, err
	}
	m.msgStore.Reset()
	// Need to reset messagepool here after calling restart on the replicas

	runObj := &types.RunObj{
		Ch: make(chan bool, 1),
	}
	// m.restartPeers()

	m.runLock.Lock()
	m.run = run
	m.runObj = runObj
	m.runLock.Unlock()

	m.workloadInjector.InjectWorkLoad()

	return runObj, nil
}

// StopRun implements AlgoDriver
func (m *CommonDriver) StopRun() {
	m.restartPeers()
}

func (master *CommonDriver) dispatchMessage(msg *types.Message) {
	peer, ok := master.ctx.Replicas.GetReplica(msg.To)

	if ok && master.msgStore.Exists(msg.ID) {
		master.ctx.Publish(types.UnInterceptedMessage, &types.MessageWrapper{
			Msg: msg,
		})

		master.sendPeerMsg(peer, msg)
	}
}

func (m *CommonDriver) dispatchTimeout(to types.ReplicaID, t string) {
	peer, ok := m.ctx.Replicas.GetReplica(to)
	if ok {
		m.sendPeerTimeout(peer, t)
	}
}

// func (m *CommonDriver) checkPeer(peer *types.Replica) {
// 	resp, err := transport.SendMsg(http.MethodGet, peer.Addr+"/health", "")
// 	if (err != nil && err.IsCode(transport.ErrBadResponse)) || resp != "Ok!" {
// 		log.Debug(
// 			fmt.Sprintf("Could not connect to peer: %#v", peer),
// 		)
// 	}
// }

func (m *CommonDriver) handleIncoming(event types.ContextEvent) {

	if event.Type != types.InterceptedMessage {
		return
	}
	msgW, ok := event.Data.(*types.MessageWrapper)
	if !ok {
		m.logger.With(map[string]interface{}{
			"msg":        "invalid event data",
			"event_type": event.Type.String(),
		})
		return
	}

	m.logger.With(map[string]interface{}{
		"intercepted_message": fmt.Sprintf("%#v", msgW.Msg),
		"run":                 strconv.Itoa(msgW.Run),
	}).Debug("Received message")
	msg := msgW.Msg
	m.msgStore.Add(msg)
	if msg.Intercept {
		m.ctx.Publish(types.ScheduledMessage, msgW)
	} else {
		m.msgStore.Mark(msg.ID)
	}
}

func (m *CommonDriver) handleOutgoing(event types.ContextEvent) {
	if event.Type != types.EnabledMessage {
		return
	}
	msgW, ok := event.Data.(*types.MessageWrapper)
	if !ok {
		return
	}
	msg := msgW.Msg
	if msg.Timeout {
		go m.dispatchTimeout(msg.To, msg.Type)
	} else {
		m.msgStore.Mark(msg.ID)
	}
}

func (m *CommonDriver) poll() {
	for {
		select {
		case event := <-m.messagesIn:
			go m.handleIncoming(event)
		case event := <-m.fromEngine:
			go m.handleOutgoing(event)
		case peer := <-m.updateCh:
			if msg, err := m.msgStore.FetchOne(peer); err == nil {
				go m.dispatchMessage(msg)
			}
		case <-m.stopCh:
			return
		}
	}
}

func (m *CommonDriver) sendPeerMsg(peer *types.Replica, msg *types.Message) {
	b, err := json.Marshal(msg)
	if err != nil {
		return
	}
	// logger.Debug(fmt.Sprintf("Sending peer message: %s", string(b)))
	transport.SendMsg(http.MethodPost, peer.Addr+"/message", string(b), transport.JsonRequest())
}

func (m *CommonDriver) sendPeerDirective(peer *types.Replica, d *DirectiveMessage) (string, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return "", err
	}
	return transport.SendMsg(http.MethodPost, peer.Addr+"/directive", string(b), transport.JsonRequest())
}

func (m *CommonDriver) sendPeerTimeout(peer *types.Replica, t string) {
	timeout := &types.Timeout{
		Type: t,
	}
	b, err := json.Marshal(timeout)
	if err != nil {
		return
	}
	transport.SendMsg(http.MethodPost, peer.Addr+"/timeout", string(b), transport.JsonRequest())
}

// Stop implements AlgoDriver
func (m *CommonDriver) Stop() {
	close(m.stopCh)
}
