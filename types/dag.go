package types

import (
	"sync"
)

type NodeSet struct {
	nodes map[uint]*Node
	lock  *sync.Mutex
}

func NewNodeSet() *NodeSet {
	return &NodeSet{
		nodes: make(map[uint]*Node),
		lock:  new(sync.Mutex),
	}
}

func (s *NodeSet) Add(n *Node) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.nodes[n.ID] = n
}

func (s *NodeSet) Get(id uint) *Node {
	s.lock.Lock()
	n, ok := s.nodes[id]
	s.lock.Unlock()
	if !ok {
		return nil
	}
	return n
}

func (s *NodeSet) Iter() []*Node {
	result := make([]*Node, len(s.nodes))
	i := 0
	s.lock.Lock()
	for _, n := range s.nodes {
		result[i] = n
		i++
	}
	s.lock.Unlock()
	return result
}

func (s *NodeSet) Union(n *NodeSet) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, node := range n.Iter() {
		s.nodes[node.ID] = node
	}
}

func (s *NodeSet) Has(id uint) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, ok := s.nodes[id]
	return ok
}

type Node struct {
	ID          uint
	Event       *Event
	Parents     *NodeSet
	Children    *NodeSet
	Ancestors   *NodeSet
	Descendants *NodeSet
	lock        *sync.Mutex
}

func NewNode(e *Event, parents []*Node) *Node {
	n := &Node{
		ID:          e.ID,
		Event:       e,
		Parents:     NewNodeSet(),
		Children:    NewNodeSet(),
		Ancestors:   NewNodeSet(),
		Descendants: NewNodeSet(),
		lock:        new(sync.Mutex),
	}

	for _, p := range parents {
		n.Parents.Add(p)
		p.Children.Add(n)
		n.Ancestors.Add(p)
		n.Ancestors.Union(p.Ancestors)
	}

	for _, a := range n.Ancestors.Iter() {
		a.Descendants.Add(n)
	}
	return n
}

func (n *Node) IsDescendant(o *Node) bool {
	return n.Descendants.Has(o.ID)
}

func (n *Node) Le(o *Node) bool {
	if o.ID == n.ID {
		return true
	}
	return n.IsDescendant(o)
}

func (n *Node) Lt(o *Node) bool {
	return n.IsDescendant(o)
}

func (n *Node) Eq(o *Node) bool {
	return n.ID == o.ID
}

func (n *Node) Clone() *Node {
	n.lock.Lock()
	defer n.lock.Unlock()
	return &Node{
		ID:          n.ID,
		Event:       n.Event,
		Parents:     NewNodeSet(),
		Children:    NewNodeSet(),
		Ancestors:   NewNodeSet(),
		Descendants: NewNodeSet(),
		lock:        new(sync.Mutex),
	}
}

func (n *Node) UpdateEvent(e *Event) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.Event = e
	n.ID = e.ID
}

type EventGraph struct {
	Nodes *NodeSet
	Roots *NodeSet
}

func NewEventGraph() *EventGraph {
	return &EventGraph{
		Nodes: NewNodeSet(),
		Roots: NewNodeSet(),
	}
}

func (g *EventGraph) AddEvent(e *Event) error {
	return nil
}
