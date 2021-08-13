package timeout

import (
	"fmt"
	"sync"

	"github.com/ds-test-framework/scheduler/types"
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

type Edge struct {
	from   *types.Node
	to     *types.Node
	weight int
}

func NewEdge(from, to *types.Node, w int) *Edge {
	return &Edge{
		from:   from,
		to:     to,
		weight: w,
	}
}

type DirectedGraph struct {
	nodes map[uint]*types.Node
	edges map[uint]map[uint]*Edge
	roots map[uint]*types.Node
	lock  *sync.Mutex
}

func NewDirectedGraph() *DirectedGraph {
	return &DirectedGraph{
		nodes: make(map[uint]*types.Node),
		edges: make(map[uint]map[uint]*Edge),
		roots: make(map[uint]*types.Node),
		lock:  new(sync.Mutex),
	}
}

func (g *DirectedGraph) GetNode(id uint) (*types.Node, bool) {
	g.lock.Lock()
	defer g.lock.Unlock()
	node, ok := g.nodes[id]
	if !ok {
		return nil, false
	}
	return node, true
}

func (g *DirectedGraph) UpdateEvent(e *types.Event) {
	g.lock.Lock()
	defer g.lock.Unlock()
	n, ok := g.nodes[e.ID]
	if !ok {
		return
	}
	n.UpdateEvent(e)
}

func (g *DirectedGraph) AddEvent(e *types.Event, parents []*types.Event, weights []int) *types.Error {
	if len(parents) != len(weights) {
		return types.NewError(
			ErrWeightsMismatch,
			"Weights provided doesn't match the number of parents",
		)
	}
	parentNodes := make([]*types.Node, len(parents))
	for i, p := range parents {
		g.lock.Lock()
		pNode, ok := g.nodes[p.ID]
		g.lock.Unlock()
		if !ok {
			return types.NewError(
				ErrInvalidNodeID,
				fmt.Sprintf("Parent does not exist: %d", p.ID),
			)
		}
		parentNodes[i] = pNode
	}
	n := types.NewNode(e, parentNodes)
	g.lock.Lock()
	g.nodes[n.ID] = n
	for i, pNode := range parentNodes {
		if _, ok := g.edges[pNode.ID]; !ok {
			g.edges[pNode.ID] = make(map[uint]*Edge)
		}
		g.edges[pNode.ID][n.ID] = NewEdge(pNode, n, weights[i])
	}
	g.lock.Unlock()
	if len(parents) == 0 {
		g.lock.Lock()
		g.roots[n.ID] = n
		g.lock.Unlock()
	}
	return nil
}

type nodeStack struct {
	nodes []*types.Node
	len   int
}

func newNodeStack() *nodeStack {
	return &nodeStack{
		nodes: make([]*types.Node, 0),
		len:   0,
	}
}

func (s *nodeStack) Push(n *types.Node) {
	s.nodes = append(s.nodes, n)
	s.len = s.len + 1
}

func (s *nodeStack) Pop() *types.Node {
	last := s.nodes[s.len-1]
	s.nodes = s.nodes[:s.len-1]
	s.len = s.len - 1
	return last
}

func (s *nodeStack) Last() *types.Node {
	return s.nodes[s.len-1]
}

func (s *nodeStack) Empty() bool {
	return s.len == 0
}

func (g *DirectedGraph) topologicalSortBetween(start, end *types.Node) []*types.Node {
	out := make([]*types.Node, 0)
	stack := newNodeStack()
	visited := make(map[uint]bool)

	stack.Push(start)
	for !stack.Empty() {
		cur := stack.Last()
		visited[cur.ID] = true
		allChildrenVisted := true
		if cur.Lt(end) {
			for _, c := range cur.Children.Iter() {
				_, ok := visited[c.ID]
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

func reverse(l []*types.Node) []*types.Node {
	length := len(l)
	res := make([]*types.Node, length)
	for i, e := range l {
		res[length-i-1] = e
	}
	return res
}

type Traverser = func(*types.Event, []*types.Event) ([]*types.Event, *types.Error)

func (g *DirectedGraph) FindPath(start, end *types.Event, traverser Traverser) ([]*types.Event, int, *types.Error) {

	// logger.Debug(fmt.Sprintf("Graph: Finding path between: %#v %#v", start, end))
	path := make([]*types.Event, 0)
	stack := newNodeStack()

	g.lock.Lock()
	defer g.lock.Unlock()
	startN, sok := g.nodes[start.ID]
	endN, eok := g.nodes[end.ID]
	if !sok || !eok {
		return path, 0, types.NewError(
			ErrInvalidNodeID,
			"Start or end event does not exist",
		)
	}
	cur := startN
	replica := start.Replica
	stack.Push(startN)
	for !cur.Eq(endN) {
		if stack.Empty() {
			return path, 0, types.NewError(
				ErrTraversalEnd,
				"Traversal ended without reaching the end node",
			)
		}
		cur = stack.Pop()
		if endN.Le(cur) {
			continue
		}
		path = append(path, cur.Event)
		// logger.Debug(fmt.Sprintf("Graph: Adding to path: %#v", cur.Event))
		filteredChildren := make([]*types.Event, 0)
		for _, c := range cur.Children.Iter() {
			if c.Event.Replica == replica {
				filteredChildren = append(filteredChildren, c.Event)
			}
		}
		traverserOut, err := traverser(cur.Event, filteredChildren)
		if err != nil {
			return path, 0, err
		}
		for _, e := range traverserOut {
			n, ok := g.nodes[e.ID]
			if !ok {
				return path, 0, types.NewError(
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
		nodes: make(map[uint]*types.Node),
		edges: make(map[uint]map[uint]*Edge),
		roots: make(map[uint]*types.Node),
		lock:  new(sync.Mutex),
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	for id, n := range g.nodes {
		d.nodes[id] = n.Clone()
	}

	for id, n := range g.nodes {
		dNode := d.nodes[id]
		for _, p := range n.Parents.Iter() {
			dNode.Parents.Add(d.nodes[p.ID])
		}
		for _, c := range n.Children.Iter() {
			dNode.Children.Add(d.nodes[c.ID])
		}
		for _, a := range n.Ancestors.Iter() {
			dNode.Ancestors.Add(d.nodes[a.ID])
		}
		for _, de := range n.Descendants.Iter() {
			dNode.Descendants.Add(d.nodes[de.ID])
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

func (g *DirectedGraph) HeaviestPath(start, end *types.Event) ([]*types.Event, int, *types.Error) {
	path := make([]*types.Event, 0)
	prev := make(map[uint]*types.Node)
	weights := make(map[uint]int)

	g.lock.Lock()
	defer g.lock.Unlock()
	startN, sok := g.nodes[start.ID]
	endN, eok := g.nodes[end.ID]
	if !sok || !eok {
		return path, 0, types.NewError(
			ErrInvalidNodeID,
			"Start or end event does not exist",
		)
	}
	order := g.topologicalSortBetween(startN, endN)

	for _, n := range order {
		weights[n.ID] = MinInt
		prev[n.ID] = nil
	}
	weights[startN.ID] = 0
	for _, n := range order {
		curWeight := weights[n.ID]
		for _, c := range n.Children.Iter() {
			if _, ok := g.edges[n.ID]; ok {
				if edge, ok := g.edges[n.ID][c.ID]; ok {
					if weights[c.ID] < curWeight+edge.weight {
						prev[c.ID] = n
						weights[c.ID] = curWeight + edge.weight
					}
				} else {
					return path, 0, types.NewError(
						ErrNoEdge,
						"No edge",
					)
				}
			} else {
				return path, 0, types.NewError(
					ErrNoEdge,
					"No edge",
				)
			}
		}
	}

	nPath := make([]*types.Node, 0)
	cur := endN
	for !cur.Eq(startN) {
		nPath = append(nPath, cur)
		cur = prev[cur.ID]
	}
	for _, n := range nPath {
		path = append(path, n.Event)
	}

	return path, weights[end.ID], nil
}
