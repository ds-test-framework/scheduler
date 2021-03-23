package types

import (
	"fmt"
	"sync"
)

const (
	ErrInvalidNodeID   = "INVALID_NODE_ID"
	ErrWeightsMismatch = "WEIGHTS_MISMATCH"
	ErrTraversalEnd    = "TRAVERSAL_ENDED"
	ErrInvalidEventID  = "INVALID_EVENT_ID"
	ErrNoEdge          = "NO_EDGE"

	MaxInt = int(^uint(0) >> 1)
	MinInt = -MaxInt - 1
)

type nodeSet struct {
	nodes map[uint]*Node
	lock  *sync.Mutex
}

func newNodeSet() *nodeSet {
	return &nodeSet{
		nodes: make(map[uint]*Node),
		lock:  new(sync.Mutex),
	}
}

func (s *nodeSet) Add(n *Node) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.nodes[n.id] = n
}

func (s *nodeSet) Get(id uint) *Node {
	s.lock.Lock()
	n, ok := s.nodes[id]
	s.lock.Unlock()
	if !ok {
		return nil
	}
	return n
}

func (s *nodeSet) Iter() []*Node {
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

func (s *nodeSet) Union(n *nodeSet) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, node := range n.Iter() {
		s.nodes[node.id] = node
	}
}

func (s *nodeSet) Has(id uint) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, ok := s.nodes[id]
	return ok
}

type Node struct {
	id          uint
	event       *Event
	parents     *nodeSet
	children    *nodeSet
	ancestors   *nodeSet
	descendants *nodeSet
	lock        *sync.Mutex
}

func NewNode(e *Event, parents []*Node) *Node {
	n := &Node{
		id:          e.ID,
		event:       e,
		parents:     newNodeSet(),
		children:    newNodeSet(),
		ancestors:   newNodeSet(),
		descendants: newNodeSet(),
		lock:        new(sync.Mutex),
	}

	for _, p := range parents {
		n.parents.Add(p)
		p.children.Add(n)
		n.ancestors.Add(p)
		n.ancestors.Union(p.ancestors)
	}

	for _, a := range n.ancestors.Iter() {
		a.descendants.Add(n)
	}
	return n
}

func (n *Node) IsDescendant(o *Node) bool {
	return n.descendants.Has(o.id)
}

func (n *Node) Le(o *Node) bool {
	if o.id == n.id {
		return true
	}
	return n.IsDescendant(o)
}

func (n *Node) Lt(o *Node) bool {
	return n.IsDescendant(o)
}

func (n *Node) Eq(o *Node) bool {
	return n.id == o.id
}

func (n *Node) Clone() *Node {
	n.lock.Lock()
	defer n.lock.Unlock()
	return &Node{
		id:          n.id,
		event:       n.event,
		parents:     newNodeSet(),
		children:    newNodeSet(),
		ancestors:   newNodeSet(),
		descendants: newNodeSet(),
		lock:        new(sync.Mutex),
	}
}

func (n *Node) UpdateEvent(e *Event) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.event = e
}

type Edge struct {
	from   *Node
	to     *Node
	weight int
}

func NewEdge(from, to *Node, w int) *Edge {
	return &Edge{
		from:   from,
		to:     to,
		weight: w,
	}
}

type DirectedGraph struct {
	nodes map[uint]*Node
	edges map[uint]map[uint]*Edge
	roots map[uint]*Node
	lock  *sync.Mutex
}

func NewDirectedGraph() *DirectedGraph {
	return &DirectedGraph{
		nodes: make(map[uint]*Node),
		edges: make(map[uint]map[uint]*Edge),
		roots: make(map[uint]*Node),
		lock:  new(sync.Mutex),
	}
}

func (g *DirectedGraph) GetNode(id uint) (*Node, bool) {
	g.lock.Lock()
	defer g.lock.Unlock()
	node, ok := g.nodes[id]
	if !ok {
		return nil, false
	}
	return node, true
}

func (g *DirectedGraph) UpdateEvent(e *Event) {
	g.lock.Lock()
	defer g.lock.Unlock()
	n, ok := g.nodes[e.ID]
	if !ok {
		return
	}
	n.UpdateEvent(e)
}

func (g *DirectedGraph) AddEvent(e *Event, parents []*Event, weights []int) *Error {
	if len(parents) != len(weights) {
		return NewError(
			ErrWeightsMismatch,
			"Weights provided doesn't match the number of parents",
		)
	}
	parentNodes := make([]*Node, len(parents))
	for i, p := range parents {
		g.lock.Lock()
		pNode, ok := g.nodes[p.ID]
		g.lock.Unlock()
		if !ok {
			return NewError(
				ErrInvalidNodeID,
				fmt.Sprintf("Parent does not exist: %d", p.ID),
			)
		}
		parentNodes[i] = pNode
	}
	n := NewNode(e, parentNodes)
	g.lock.Lock()
	g.nodes[n.id] = n
	for i, pNode := range parentNodes {
		if _, ok := g.edges[pNode.id]; !ok {
			g.edges[pNode.id] = make(map[uint]*Edge)
		}
		g.edges[pNode.id][n.id] = NewEdge(pNode, n, weights[i])
	}
	g.lock.Unlock()
	if len(parents) == 0 {
		g.lock.Lock()
		g.roots[n.id] = n
		g.lock.Unlock()
	}
	return nil
}

type nodeStack struct {
	nodes []*Node
	len   int
}

func newNodeStack() *nodeStack {
	return &nodeStack{
		nodes: make([]*Node, 0),
		len:   0,
	}
}

func (s *nodeStack) Push(n *Node) {
	s.nodes = append(s.nodes, n)
	s.len = s.len + 1
}

func (s *nodeStack) Pop() *Node {
	last := s.nodes[s.len-1]
	s.nodes = s.nodes[:s.len-1]
	s.len = s.len - 1
	return last
}

func (s *nodeStack) Last() *Node {
	return s.nodes[s.len-1]
}

func (s *nodeStack) Empty() bool {
	return s.len == 0
}

func (g *DirectedGraph) topologicalSortBetween(start, end *Node) []*Node {
	out := make([]*Node, 0)
	stack := newNodeStack()
	visited := make(map[uint]bool)

	stack.Push(start)
	for !stack.Empty() {
		cur := stack.Last()
		visited[cur.id] = true
		allChildrenVisted := true
		if cur.Lt(end) {
			for _, c := range cur.children.Iter() {
				_, ok := visited[c.id]
				if !ok {
					allChildrenVisted = false
					stack.Push(c)
				}
			}
		}
		if allChildrenVisted {
			out = append(out, stack.Pop())
		}
	}
	return reverse(out)
}

func reverse(l []*Node) []*Node {
	length := len(l)
	res := make([]*Node, length)
	for i, e := range l {
		res[length-i-1] = e
	}
	return res
}

type Traverser = func(*Event, []*Event) ([]*Event, *Error)

func (g *DirectedGraph) FindPath(start, end *Event, traverser Traverser) ([]*Event, int, *Error) {

	// logger.Debug(fmt.Sprintf("Graph: Finding path between: %#v %#v", start, end))
	path := make([]*Event, 0)
	stack := newNodeStack()

	g.lock.Lock()
	defer g.lock.Unlock()
	startN, sok := g.nodes[start.ID]
	endN, eok := g.nodes[end.ID]
	if !sok || !eok {
		return path, 0, NewError(
			ErrInvalidNodeID,
			"Start or end event does not exist",
		)
	}
	cur := startN
	replica := start.Replica
	stack.Push(startN)
	for !cur.Eq(endN) {
		if stack.Empty() {
			return path, 0, NewError(
				ErrTraversalEnd,
				"Traversal ended without reaching the end node",
			)
		}
		cur = stack.Pop()
		if endN.Le(cur) {
			continue
		}
		path = append(path, cur.event)
		// logger.Debug(fmt.Sprintf("Graph: Adding to path: %#v", cur.event))
		filteredChildren := make([]*Event, 0)
		for _, c := range cur.children.Iter() {
			if c.event.Replica == replica {
				filteredChildren = append(filteredChildren, c.event)
			}
		}
		traverserOut, err := traverser(cur.event, filteredChildren)
		if err != nil {
			return path, 0, err
		}
		for _, e := range traverserOut {
			n, ok := g.nodes[e.ID]
			if !ok {
				return path, 0, NewError(
					ErrInvalidNodeID,
					"Invalid node ID when traversing",
				)
			}
			stack.Push(n)
		}
	}

	weight := 0
	for i := 1; i < len(path); i++ {
		if edgeStart, ok := g.edges[path[i-1].ID]; ok {
			if edge, ok := edgeStart[path[i].ID]; ok {
				weight = weight + edge.weight
			}
		}
	}
	return path, weight, nil
}

func (g *DirectedGraph) Clone() *DirectedGraph {
	d := &DirectedGraph{
		nodes: make(map[uint]*Node),
		edges: make(map[uint]map[uint]*Edge),
		roots: make(map[uint]*Node),
		lock:  new(sync.Mutex),
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	for id, n := range g.nodes {
		d.nodes[id] = n.Clone()
	}

	for id, n := range g.nodes {
		dNode := d.nodes[id]
		for _, p := range n.parents.Iter() {
			dNode.parents.Add(d.nodes[p.id])
		}
		for _, c := range n.children.Iter() {
			dNode.children.Add(d.nodes[c.id])
		}
		for _, a := range n.ancestors.Iter() {
			dNode.ancestors.Add(d.nodes[a.id])
		}
		for _, de := range n.descendants.Iter() {
			dNode.descendants.Add(d.nodes[de.id])
		}
	}

	for f, m := range g.edges {
		d.edges[f] = make(map[uint]*Edge)
		for t, e := range m {
			d.edges[f][t] = NewEdge(d.nodes[f], d.nodes[t], e.weight)
		}
	}

	for id := range g.roots {
		d.roots[id] = d.nodes[id]
	}

	return d
}

func (g *DirectedGraph) HeaviestPath(start, end *Event) ([]*Event, int, *Error) {
	path := make([]*Event, 0)
	prev := make(map[uint]*Node)
	weights := make(map[uint]int)

	g.lock.Lock()
	defer g.lock.Unlock()
	startN, sok := g.nodes[start.ID]
	endN, eok := g.nodes[end.ID]
	if !sok || !eok {
		return path, 0, NewError(
			ErrInvalidNodeID,
			"Start or end event does not exist",
		)
	}
	order := g.topologicalSortBetween(startN, endN)

	for _, n := range order {
		weights[n.id] = MinInt
		prev[n.id] = nil
	}
	weights[startN.id] = 0
	for _, n := range order {
		curWeight := weights[n.id]
		for _, c := range n.children.Iter() {
			if _, ok := g.edges[n.id]; ok {
				if edge, ok := g.edges[n.id][c.id]; ok {
					if weights[c.id] < curWeight+edge.weight {
						prev[c.id] = n
						weights[c.id] = curWeight + edge.weight
					}
				} else {
					return path, 0, NewError(
						ErrNoEdge,
						"No edge",
					)
				}
			} else {
				return path, 0, NewError(
					ErrNoEdge,
					"No edge",
				)
			}
		}
	}

	nPath := make([]*Node, 0)
	cur := endN
	for !cur.Eq(startN) {
		nPath = append(nPath, cur)
		cur = prev[cur.id]
	}
	for _, n := range nPath {
		path = append(path, n.event)
	}

	return path, weights[end.ID], nil
}
