package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ds-test-framework/scheduler/pkg/logger"
	transport "github.com/ds-test-framework/scheduler/pkg/transports/http"
	"github.com/ds-test-framework/scheduler/pkg/types"
	"github.com/spf13/viper"
)

const (
	ErrPeersNotReady = "ERR_PEERS_NOT_READY"
)

// Replica contains information of a replica
type Replica struct {
	ID    types.ReplicaID        `json:"id"`
	Addr  string                 `json:"addr"`
	Info  map[string]interface{} `json:"info,omitempty"`
	Ready bool                   `json:"ready"`
}

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
	toEngine         chan *types.MessageWrapper
	fromEngine       chan *types.MessageWrapper
	workloadInjector WorkloadInjector

	peers       *PeerStore
	totalPeers  int
	msgStore    *msgStore
	transport   *transport.HttpTransport
	transportIn chan string

	counter     int
	counterLock *sync.Mutex

	stopCh   chan bool
	updateCh chan types.ReplicaID

	run     int
	runObj  *types.RunObj
	runLock *sync.Mutex
}

// NewCommonDriver constructs a CommonDriver
// Config options expected
// - type: common
// - algo: for fetching the workload injector of the specific algorithm
// - transport: transport config options
// - num_replicas: number of replicas that are used for testing
func NewCommonDriver(c *viper.Viper, workloadInjector WorkloadInjector) *CommonDriver {
	t := transport.NewHttpTransport(c.Sub("transport"))
	n := &CommonDriver{
		toEngine:         make(chan *types.MessageWrapper, 10),
		fromEngine:       make(chan *types.MessageWrapper, 10),
		peers:            NewPeerStore(),
		totalPeers:       c.GetInt("num_replicas"),
		workloadInjector: workloadInjector,
		counter:          0,
		counterLock:      new(sync.Mutex),
		stopCh:           make(chan bool),
		updateCh:         make(chan types.ReplicaID, 2),
		transport:        t,
		transportIn:      t.ReceiveChan(),

		run:     0,
		runLock: new(sync.Mutex),
	}

	n.msgStore = newMsgStore(n.updateCh)
	return n
}

// Init implements AlgoDriver
func (m *CommonDriver) Init() {
	go m.transport.Run()
	go m.poll()
}

// InChan implements AlgoDriver
func (m *CommonDriver) InChan() chan *types.MessageWrapper {
	return m.fromEngine
}

// OutChan implements AlgoDriver
func (m *CommonDriver) OutChan() chan *types.MessageWrapper {
	return m.toEngine
}

func (m *CommonDriver) waitForAllPeers() *types.Error {
	timer := time.After(10 * time.Second)
	for {
		select {
		case <-timer:
			return types.NewError(
				ErrPeersNotReady,
				"All peers not connected",
			)
		default:
		}
		if m.peers.Count() != m.totalPeers {
			continue
		}
		for _, p := range m.peers.Iter() {
			if !p.Ready {
				continue
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

		if m.peers.Count() == m.totalPeers {
			return true, nil
		}
	}
}

func (m *CommonDriver) restartPeers() {
	var wg sync.WaitGroup
	for _, p := range m.peers.Iter() {
		wg.Add(1)
		go func(peer *Replica) {
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

func (master *CommonDriver) dispatchMessage(msg *InterceptedMessage) {
	peer, ok := master.peers.GetPeer(msg.To)

	if ok && master.msgStore.Exists(msg.ID) {
		master.sendPeerMsg(peer, msg)
	}
}

func (m *CommonDriver) dispatchTimeout(to types.ReplicaID, t string) {
	peer, ok := m.peers.GetPeer(to)
	if ok {
		m.sendPeerTimeout(peer, t)
	}
}

func (m *CommonDriver) checkPeer(peer *Replica) {
	resp, err := m.transport.SendMsg(http.MethodGet, peer.Addr+"/health", "")
	if (err != nil && err.IsCode(transport.ErrBadResponse)) || resp != "Ok!" {
		logger.Debug(
			fmt.Sprintf("Could not connect to peer: %#v", peer),
		)
	}
}

func (m *CommonDriver) handleIncoming(msg string) {
	logger.Debug("Received message: " + msg)
	r, err := unmarshal(msg)
	if err != nil {
		logger.Debug("Error unmarshalling message: " + err.Error())
		return
	}
	switch r.Type {
	case requestMessage:
		m.counterLock.Lock()
		id := m.counter
		m.counter += 1
		m.counterLock.Unlock()

		iMsg := r.Message
		iMsg.ID = strconv.Itoa(id)

		m.msgStore.Add(&iMsg)

		if iMsg.Intercept {
			m.runLock.Lock()
			run := m.run
			m.runLock.Unlock()
			go func() {
				// logger.Debug("Sending message to engine: " + msg)
				m.toEngine <- &types.MessageWrapper{
					Run: run,
					Msg: &types.Message{
						Type:    iMsg.T,
						From:    types.ReplicaID(iMsg.From),
						To:      types.ReplicaID(iMsg.To),
						ID:      iMsg.ID,
						Msg:     []byte(iMsg.Msg),
						Timeout: false,
						Weight:  0,
					},
				}
			}()
		} else {
			m.msgStore.Mark(iMsg.ID)
		}
	case requestPeerRegister:
		logger.Debug(fmt.Sprintf("Peer connected: %#v", r.Peer))
		m.peers.AddPeer(r.Peer)
		go m.checkPeer(r.Peer)
	case timeoutMessage:
		m.runLock.Lock()
		run := m.run
		m.runLock.Unlock()

		m.counterLock.Lock()
		id := m.counter
		m.counter += 1
		m.counterLock.Unlock()

		go func() {
			m.toEngine <- &types.MessageWrapper{
				Run: run,
				Msg: &types.Message{
					Type:    r.Timeout.Type,
					ID:      strconv.Itoa(id),
					From:    types.ReplicaID(r.Timeout.Peer),
					To:      types.ReplicaID(r.Timeout.Peer),
					Msg:     []byte{},
					Weight:  r.Timeout.Duration,
					Timeout: true,
				},
			}
		}()
	}
}

func (m *CommonDriver) poll() {
	for {
		select {
		case req := <-m.transportIn:
			go m.handleIncoming(req)
		case msgIn := <-m.fromEngine:
			if msgIn.Msg.Timeout {
				go m.dispatchTimeout(msgIn.Msg.To, msgIn.Msg.Type)
			} else {
				m.msgStore.Mark(msgIn.Msg.ID)
			}
		case peer := <-m.updateCh:
			if msg, err := m.msgStore.FetchOne(peer); err == nil {
				go m.dispatchMessage(msg)
			}
		case <-m.stopCh:
			return
		}
	}
}

func (m *CommonDriver) sendPeerMsg(peer *Replica, msg *InterceptedMessage) {
	b, err := json.Marshal(msg)
	if err != nil {
		return
	}
	logger.Debug(fmt.Sprintf("Sending peer message: %s", string(b)))
	m.transport.SendMsg(http.MethodPost, peer.Addr+"/message", string(b), transport.JsonRequest())
}

func (m *CommonDriver) sendPeerDirective(peer *Replica, d *DirectiveMessage) (string, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return "", err
	}
	return m.transport.SendMsg(http.MethodPost, peer.Addr+"/directive", string(b), transport.JsonRequest())
}

func (m *CommonDriver) sendPeerTimeout(peer *Replica, t string) {
	timeout := &timeout{
		Type: t,
	}
	b, err := json.Marshal(timeout)
	if err != nil {
		return
	}
	m.transport.SendMsg(http.MethodPost, peer.Addr+"/timeout", string(b), transport.JsonRequest())
}

// Destroy implements AlgoDriver
func (m *CommonDriver) Destroy() {
	close(m.stopCh)
	m.transport.Stop()
}
