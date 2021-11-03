package types

import (
	"encoding/json"
	"sync"
)

// EventNode is a node of the event acyclic directed graph
type EventNode struct {
	Event *Event `json:"event"`
	prev  *EventNode
	next  *EventNode
	// Parents is the set of parent nodes
	Parents *EventNodeSet `json:"parents"`
	// Children is the set of child nodes
	Children *EventNodeSet `json:"children"`
	dirty    bool
	lock     *sync.Mutex
}

// NewEventNode instantiates a new EventNode
func NewEventNode(e *Event) *EventNode {
	return &EventNode{
		Event:    e,
		prev:     nil,
		next:     nil,
		Parents:  NewEventNodeSet(),
		Children: NewEventNodeSet(),
		dirty:    false,
		lock:     new(sync.Mutex),
	}
}

// SetPrev sets the prev attribute (thread safe) in  the current node's strand
func (n *EventNode) SetPrev(prev *EventNode) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.prev = prev
}

// GetPrev returns the previous node in the strand (prev attribute)
func (n *EventNode) GetPrev() *EventNode {
	n.lock.Lock()
	defer n.lock.Unlock()
	return n.prev
}

// SetNext sets the next attibute (thread safe)
func (n *EventNode) SetNext(next *EventNode) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.next = next
}

// GetNext returns the next node of the current node's strand
func (n *EventNode) GetNext() *EventNode {
	n.lock.Lock()
	defer n.lock.Unlock()
	return n.next
}

func (n *EventNode) MarkDirty() {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.dirty = true
}

func (n *EventNode) MarkClean() {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.dirty = false
}

// AddParents updates Parents, Ancestors of the current node and
// Children, Descendants of the parent and ancestor nodes respectively
func (n *EventNode) AddParents(parents []*EventNode) {
	for _, parent := range parents {
		n.Parents.Add(parent)
		parent.Children.Add(n)
	}
}

func (n *EventNode) MarshalJSON() ([]byte, error) {
	keyvals := make(map[string]interface{})
	keyvals["event"] = n.Event
	keyvals["parents"] = n.Parents
	keyvals["children"] = n.Children
	keyvals["prev"] = nil
	keyvals["next"] = nil

	n.lock.Lock()
	if n.prev != nil {
		keyvals["prev"] = n.prev.Event.ID
	}
	if n.next != nil {
		keyvals["next"] = n.next.Event.ID
	}
	n.lock.Unlock()

	return json.Marshal(keyvals)
}

// EventDAG is the directed acyclic graph of events
type EventDAG struct {
	nodes       *EventNodeSet
	strands     map[ReplicaID]*EventNode
	latestNodes map[ReplicaID]*EventNode
	lock        *sync.Mutex
}

// NewEventDag instantiates EventDAG
func NewEventDag() *EventDAG {
	return &EventDAG{
		nodes:       NewEventNodeSet(),
		strands:     make(map[ReplicaID]*EventNode),
		latestNodes: make(map[ReplicaID]*EventNode),
		lock:        new(sync.Mutex),
	}
}

// AddNode adds an event with the respective parent nodes
func (d *EventDAG) AddNode(e *Event, parents []*Event) {
	parentNodes := make([]*EventNode, 0)
	for _, parent := range parents {
		if d.nodes.Has(parent.ID) {
			parentNodes = append(parentNodes, d.nodes.Get(parent.ID))
		}
	}

	node := NewEventNode(e)
	l, ok := d.latestNodes[e.Replica]
	if ok {
		l.SetNext(node)
		node.SetPrev(l)
		d.latestNodes[e.Replica] = node
		parentNodes = append(parentNodes, l)
	} else {
		d.latestNodes[e.Replica] = node
	}

	node.AddParents(parentNodes)

	d.nodes.Add(node)

	d.lock.Lock()
	defer d.lock.Unlock()
	_, ok = d.strands[e.Replica]
	if !ok {
		d.strands[e.Replica] = node
	}
}

// GetStrand returns the strand of events for that particular replica
func (d *EventDAG) GetStrand(replica ReplicaID) (*EventNode, bool) {
	d.lock.Lock()
	defer d.lock.Unlock()
	head, ok := d.strands[replica]
	return head, ok
}

// GetLatestEvent returns the latest event in that replica strand
func (d *EventDAG) GetLatestEvent(replica ReplicaID) (*EventNode, bool) {
	d.lock.Lock()
	defer d.lock.Unlock()
	e, ok := d.latestNodes[replica]
	return e, ok
}

func (d *EventDAG) MarshalJSON() ([]byte, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	keyvals := make(map[string]interface{})
	nodes := make(map[uint64]*EventNode, d.nodes.Size())

	for _, node := range d.nodes.Iter() {
		nodes[node.Event.ID] = node
	}

	keyvals["nodes"] = nodes
	keyvals["strands"] = d.strands
	return json.MarshalIndent(keyvals, "", "\t")
}
