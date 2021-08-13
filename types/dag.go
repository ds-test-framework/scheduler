package types

import "sync"

type EventNode struct {
	*Event
	prev        *EventNode
	next        *EventNode
	Parents     *EventNodeSet
	Children    *EventNodeSet
	Descendants *EventNodeSet
	Ancestors   *EventNodeSet
	lock        *sync.Mutex
}

func NewEventNode(e *Event) *EventNode {
	return &EventNode{
		Event:       e,
		prev:        nil,
		next:        nil,
		Parents:     NewEventNodeSet(),
		Children:    NewEventNodeSet(),
		Ancestors:   NewEventNodeSet(),
		Descendants: NewEventNodeSet(),
		lock:        new(sync.Mutex),
	}
}

func (n *EventNode) SetPrev(prev *EventNode) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.prev = prev
}

func (n *EventNode) Prev() *EventNode {
	n.lock.Lock()
	defer n.lock.Unlock()
	return n.prev
}

func (n *EventNode) SetNext(next *EventNode) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.next = next
}

func (n *EventNode) Next() *EventNode {
	n.lock.Lock()
	defer n.lock.Unlock()
	return n.next
}

func (n *EventNode) AddParents(parents []*EventNode) {
	for _, parent := range parents {
		n.Parents.Add(parent)
		parent.Children.Add(n)

		n.Ancestors.Add(parent)
		n.Ancestors.Union(parent.Ancestors)
	}
	for _, ancestor := range n.Ancestors.Iter() {
		ancestor.Descendants.Add(n)
	}
}

type EventDAG struct {
	nodes   *EventNodeSet
	strands map[ReplicaID]*EventNode
	lock    *sync.Mutex
}

func NewEventDag() *EventDAG {
	return &EventDAG{
		nodes:   NewEventNodeSet(),
		strands: make(map[ReplicaID]*EventNode),
		lock:    new(sync.Mutex),
	}
}

func (d *EventDAG) AddNode(e *Event, parents []*Event) {
	parentNodes := make([]*EventNode, 0)
	for _, parent := range parents {
		if d.nodes.Has(parent.ID) {
			parentNodes = append(parentNodes, d.nodes.Get(parent.ID))
		}
	}

	node := NewEventNode(e)
	node.AddParents(parentNodes)

	d.lock.Lock()
	defer d.lock.Unlock()
	_, ok := d.strands[e.Replica]
	if !ok {
		d.strands[e.Replica] = node
	}
}

func (d *EventDAG) GetStrand(replica ReplicaID) (*EventNode, bool) {
	d.lock.Lock()
	defer d.lock.Unlock()
	head, ok := d.strands[replica]
	return head, ok
}
