package timeout

import (
	"fmt"
	"sync"

	"github.com/zeu5/model-checker/pkg/types"
	"github.com/zeu5/model-checker/pkg/util"
)

const (
	ErrInvalidMsgId   = "INVALID_MSG_ID"
	ErrInvalidEventId = "INVALID_EVENT_ID"
	ErrEmptyEvent     = "EMPTY_EVENT"
)

type graphManager struct {
	latestEvents map[uint]uint
	sends        map[uint][]uint
	receives     map[uint][]uint
	graph        *types.DirectedGraph
	lock         *sync.Mutex
}

func newGraphManager() *graphManager {
	return &graphManager{
		latestEvents: make(map[uint]uint),
		sends:        make(map[uint][]uint),
		receives:     make(map[uint][]uint),
		graph:        types.NewDirectedGraph(),
		lock:         new(sync.Mutex),
	}
}

func (m *graphManager) AddEvent(
	e *types.Event,
	messages *messageStore,
	events *eventStore,
) *types.Error {
	if e == nil {
		return types.NewError(
			ErrEmptyEvent,
			"Empty event",
		)
	}
	parents := make([]*types.Event, 0)
	weights := make([]int, 0)
	prev, exists := m.updateLatest(e)
	if exists {
		prevN, ok := events.Get(prev)
		if !ok {
			return types.NewError(
				ErrInvalidEventId,
				"Previous event does not exist",
			)
		}
		parents = append(parents, prevN)
		weights = append(weights, 0)
		prevN.UpdateNext(e)
		e.UpdatePrev(prevN)
	}

	msg, ok := messages.Get(e.MsgID)
	if !ok {
		return types.NewError(
			ErrInvalidMsgId,
			"Invalid message ID",
		)
	}

	if e.Type == types.Receive {
		sendEvent, ok := events.Get(msg.GetSendEvent())
		if !ok {
			return types.NewError(
				ErrInvalidEventId,
				fmt.Sprintf("Invalid event ID: %d", msg.GetSendEvent()),
			)
		}
		parents = append(parents, sendEvent)
		weights = append(weights, msg.Weight)
	}

	if msg.Timeout {
		m.updateSendsReceives(e)
	}

	// defer logger.Debug(fmt.Sprintf("Graph Manager: Added event: %#v", e))
	return m.graph.AddEvent(e, parents, weights)
}

func (m *graphManager) updateSendsReceives(e *types.Event) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if e.Type == types.Send {
		_, ok := m.sends[e.Replica]
		if !ok {
			m.sends[e.Replica] = make([]uint, 0)
		}
		m.sends[e.Replica] = append(m.sends[e.Replica], e.ID)
	} else if e.Type == types.Receive {
		_, ok := m.receives[e.Replica]
		if !ok {
			m.receives[e.Replica] = make([]uint, 0)
		}
		m.receives[e.Replica] = append(m.receives[e.Replica], e.ID)
	}
}

func (m *graphManager) updateLatest(e *types.Event) (uint, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	cur, ok := m.latestEvents[e.Replica]
	m.latestEvents[e.Replica] = e.ID
	if !ok {
		return 0, false
	}
	return cur, true
}

func (m *graphManager) GetPseudo() *graphPseudoManager {
	g := &graphPseudoManager{
		latestEvents: make(map[uint]uint),
		sends:        make(map[uint][]uint),
		receives:     make(map[uint][]uint),
		graph:        m.graph.Clone(),
		lock:         new(sync.Mutex),
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	for id, e := range m.latestEvents {
		g.latestEvents[id] = e
	}
	for replica, sends := range m.sends {
		g.sends[replica] = sends
	}
	for replica, receives := range m.receives {
		g.receives[replica] = receives
	}
	return g
}

func (m *graphManager) Reset() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.latestEvents = make(map[uint]uint)
	m.sends = make(map[uint][]uint)
	m.receives = make(map[uint][]uint)
	m.graph = types.NewDirectedGraph()
}

type graphPseudoManager struct {
	latestEvents map[uint]uint
	sends        map[uint][]uint
	receives     map[uint][]uint
	graph        *types.DirectedGraph
	lock         *sync.Mutex
}

func (m *graphPseudoManager) AddEvent(
	e *types.Event,
	messages *messagePseudoStore,
	events *eventPseudoStore,
) (bool, *types.Error) {
	if e == nil {
		return false, types.NewError(
			ErrEmptyEvent,
			"Event is empty",
		)
	}
	// logger.Debug(fmt.Sprintf("Graph Pseudo Manager: Adding event to graph: %#v", e))
	events.MarkDirty(e.ID)
	events.Set(e)
	e, ok := events.Get(e.ID)
	if !ok {
		return false, types.NewError(
			ErrEmptyEvent,
			"Event does not exist",
		)
	}
	m.graph.UpdateEvent(e)
	messages.MarkDirty(e.MsgID)

	parents := make([]*types.Event, 0)
	weights := make([]int, 0)
	prev, exists := m.updateLatest(e)
	if exists {
		events.MarkDirty(prev)
		prevN, ok := events.Get(prev)
		if !ok {
			return false, types.NewError(
				ErrEmptyEvent,
				"Previous event does not exist",
			)
		}
		m.graph.UpdateEvent(prevN)

		parents = append(parents, prevN)
		weights = append(weights, 0)
		prevN.UpdateNext(e)
		e.UpdatePrev(prevN)
	}

	msg, ok := messages.Get(e.MsgID)
	if !ok {
		return false, types.NewError(
			ErrInvalidMsgId,
			"Invalid message ID",
		)
	}

	if e.Type == types.Receive {
		sendEvent, ok := events.Get(msg.GetSendEvent())
		if !ok {
			return false, types.NewError(
				ErrInvalidEventId,
				fmt.Sprintf("Invalid event ID: %d", msg.GetSendEvent()),
			)
		}
		parents = append(parents, sendEvent)
		weights = append(weights, msg.Weight)
	}

	if msg.Timeout {
		m.updateSendsReceives(e)
	}
	err := m.graph.AddEvent(e, parents, weights)
	if err != nil {
		return false, err
	}
	// logger.Debug(fmt.Sprintf("Graph Pseudo Manager: Added event: %#v", e))

	if msg.Timeout && e.Type == types.Receive {
		pairs := m.findPairs(e)
		// logger.Debug(fmt.Sprintf("Graph Pseudo Manager: Done finding pairs for event: %#v", e))
		for _, p := range pairs {
			pEvent, ok := events.Get(p)
			if !ok {
				return false, types.NewError(
					ErrEmptyEvent,
					"Pair event does not exist",
				)
			}
			res, err := m.checkPair(pEvent, e, events, messages)
			if err != nil {
				return false, err
			}
			if !res {
				// logger.Debug("Graph Manager: Found Conflict path!")
				return false, nil
			}
		}
	}

	return true, nil
}

func (m *graphPseudoManager) findPairs(e *types.Event) []uint {
	result := make([]uint, 0)
	m.lock.Lock()
	defer m.lock.Unlock()

	eventNode, ok := m.graph.GetNode(e.ID)
	if !ok {
		return result
	}

	for replica, receives := range m.receives {
		if replica == e.Replica {
			continue
		}
		for i := (len(receives) - 1); i >= 0; i-- {
			receiveNode, ok := m.graph.GetNode(receives[i])
			if !ok {
				continue
			}
			if receiveNode.Lt(eventNode) {
				result = append(result, receives[i])
				break
			}
		}
	}

	return result
}

func (m *graphPseudoManager) checkPair(
	r1, r2 *types.Event,
	events *eventPseudoStore,
	messages *messagePseudoStore,
) (bool, *types.Error) {
	// logger.Debug(fmt.Sprintf("Graph Pseudo Manager: Checking pair: %#v, %#v", r1, r2))

	message1, ok1 := messages.Get(r1.MsgID)
	message2, ok2 := messages.Get(r2.MsgID)

	if !ok1 || !ok2 {
		return false, types.NewError(
			ErrInvalidMsgId,
			"Invalid message id in the pair",
		)
	}
	send1, ok := events.Get(message1.GetSendEvent())
	if !ok {
		return false, types.NewError(
			ErrInvalidEventId,
			fmt.Sprintf("Invalid send1 event id: %d", send1.ID),
		)
	}
	send1N, ok := m.graph.GetNode(send1.ID)
	if !ok {
		return false, types.NewError(
			types.ErrInvalidNodeID,
			"Could not find send1 node",
		)
	}

	send2, ok := events.Get(message2.GetSendEvent())
	if !ok {
		return false, types.NewError(
			ErrInvalidEventId,
			fmt.Sprintf("Invalid send2 event id: %d", send2.ID),
		)
	}
	send2N, ok := m.graph.GetNode(send2.ID)
	if !ok {
		return false, types.NewError(
			types.ErrInvalidNodeID,
			"Could not find send2 node",
		)
	}

	// logger.Debug(fmt.Sprintf("Graph Pseudo Manager: Fetched nodes"))

	m.lock.Lock()
	sends := m.sends[r2.Replica]
	m.lock.Unlock()
	for _, send := range sends {
		sendE, _ := events.Get(send)
		sendN, ok := m.graph.GetNode(send)
		if !ok {
			return false, types.NewError(
				types.ErrInvalidNodeID,
				"Could not find send node",
			)
		}
		if sendN.Le(send2N) && sendN.Lt(send1N) {
			// logger.Debug("Graph Pseudo Manager: Found send")
			rPath, rWeight, err := m.graph.FindPath(sendE, r2, func(cur *types.Event, children []*types.Event) ([]*types.Event, *types.Error) {
				curE, ok := events.Get(cur.ID)
				if !ok {
					return children, types.NewError(
						ErrInvalidEventId,
						fmt.Sprintf("Graph pseudo manager: Invalid send event id: %d", cur.ID),
					)
				}
				message, ok := messages.Get(curE.MsgID)
				if !ok {
					return children, types.NewError(
						ErrInvalidMsgId,
						"Invalid message id in the pair",
					)
				}
				var check *types.Event
				if curE.Type == types.Send && message.Timeout {
					check, ok = events.Get(message.GetReceiveEvent())
				} else {
					check, ok = events.Get(curE.GetNext())
				}
				if !ok {
					return children, types.NewError(
						ErrInvalidEventId,
						"Graph pseudo manager: No successor",
					)
				}
				contains := false
				result := make([]*types.Event, 0)
				for _, c := range children {
					if c.ID == check.ID {
						contains = true
						continue
					}
					result = append(result, c)
				}
				if contains {
					result = append(result, check)
				}
				return result, nil
			})
			if err != nil {
				return false, err
			}
			// logger.Debug("Graph Pseudo Manager: Found main path")
			oPath, oWeight, err := m.graph.HeaviestPath(sendE, r2)
			if err != nil {
				return false, err
			}
			// logger.Debug("Graph Pseudo Manager: Comparing paths")
			if !util.PathEq(rPath, oPath) && oWeight >= rWeight {
				return false, nil
			}

			break
		}
	}
	return true, nil
}

func (m *graphPseudoManager) updateLatest(e *types.Event) (uint, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	cur, ok := m.latestEvents[e.Replica]
	m.latestEvents[e.Replica] = e.ID
	if !ok {
		return 0, false
	}
	return cur, true
}

func (m *graphPseudoManager) updateSendsReceives(e *types.Event) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if e.Type == types.Send {
		_, ok := m.sends[e.Replica]
		if !ok {
			m.sends[e.Replica] = make([]uint, 0)
		}
		m.sends[e.Replica] = append(m.sends[e.Replica], e.ID)
	} else if e.Type == types.Receive {
		_, ok := m.receives[e.Replica]
		if !ok {
			m.receives[e.Replica] = make([]uint, 0)
		}
		m.receives[e.Replica] = append(m.receives[e.Replica], e.ID)
	}
}
