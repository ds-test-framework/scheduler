package ttest

import (
	"fmt"
	"sync"

	"github.com/ds-test-framework/scheduler/log"
	"github.com/ds-test-framework/scheduler/types"
)

type Filter interface {
	Test(*ControllerMsgEnvelop) bool
}

type TypeFilter struct {
	t       string
	exclude bool
}

func NewTypeFilter(t string, exclude bool) *TypeFilter {
	return &TypeFilter{
		t:       t,
		exclude: exclude,
	}
}

func (m *TypeFilter) Test(msg *ControllerMsgEnvelop) bool {
	if msg.Type == m.t {
		return !m.exclude
	}
	return m.exclude
}

type CountTypeFilter struct {
	t         string
	maxcount  int
	count     int
	countLock *sync.Mutex
}

func NewCountTypeFilter(t string, maxcount int) *CountTypeFilter {
	return &CountTypeFilter{
		t:         t,
		maxcount:  maxcount,
		count:     0,
		countLock: new(sync.Mutex),
	}
}

func (m *CountTypeFilter) Test(msg *ControllerMsgEnvelop) bool {
	if msg.Type == m.t {
		var count int
		m.countLock.Lock()
		m.count = m.count + 1
		count = m.count
		m.countLock.Unlock()
		if count > m.maxcount {
			// log.Debug(fmt.Sprintf("CountType filter count exceeded for type: %s", m.t))
			return false
		}
	}
	return true
}

// CountTypeFilter per peer
type PeerTypeCountFilter struct {
	peers map[string]*BlockNAllow
	mtx   *sync.Mutex

	start int
	end   int
	t     string
}

func NewPeerTypeCountFilter(t string, start, end int) *PeerTypeCountFilter {
	return &PeerTypeCountFilter{
		peers: make(map[string]*BlockNAllow),
		mtx:   new(sync.Mutex),
		t:     t,
		start: start,
		end:   end,
	}
}

func (p *PeerTypeCountFilter) Test(msg *ControllerMsgEnvelop) bool {
	p.mtx.Lock()
	peerFilter, ok := p.peers[string(msg.To)]
	if !ok {
		peerFilter = NewBlockNAllow(p.t, p.start, p.end)
		p.peers[string(msg.To)] = peerFilter
	}
	p.mtx.Unlock()

	return peerFilter.Test(msg)
}

type PeerFilter struct {
	peer    string
	exclude bool
	src     bool
}

func NewPeerFilter(peer string, exclude, src bool) *PeerFilter {
	return &PeerFilter{
		peer:    peer,
		exclude: exclude,
		src:     src,
	}
}

func (p *PeerFilter) Test(msg *ControllerMsgEnvelop) bool {
	var peer string
	if p.src {
		peer = string(msg.From)
	} else {
		peer = string(msg.To)
	}
	if peer == p.peer {
		return !p.exclude
	}
	return p.exclude
}

type RandomPeerTypeFilter struct {
	peer string
	t    string
	// exclude  bool
	maxcount int

	count     int
	countLock *sync.Mutex
}

func NewRandomPeerTypeFilter(t string, maxcount int) *RandomPeerTypeFilter {
	return &RandomPeerTypeFilter{
		peer: "",
		t:    t,
		// exclude:   exclude,
		maxcount:  maxcount,
		count:     0,
		countLock: new(sync.Mutex),
	}
}

func (f *RandomPeerTypeFilter) Test(msg *ControllerMsgEnvelop) bool {
	if f.peer == "" {
		f.peer = string(msg.To)
	}
	if string(msg.To) == f.peer && msg.Type == f.t {
		if f.maxcount > 0 {
			var count int
			f.countLock.Lock()
			f.count = f.count + 1
			count = f.count
			f.countLock.Unlock()
			if count > f.maxcount {
				log.Debug(fmt.Sprintf("Count exceeded for peer: %s", f.peer))
				return false
			}
			return true
		}
		// No limit
		return false
	}
	return true
}

type BlockNAllow struct {
	t string

	startCount int
	endCount   int

	count     int
	countLock *sync.Mutex
}

func NewBlockNAllow(t string, start, end int) *BlockNAllow {
	return &BlockNAllow{
		t:          t,
		startCount: start,
		endCount:   end,
		count:      0,
		countLock:  new(sync.Mutex),
	}
}

func (b *BlockNAllow) Test(msg *ControllerMsgEnvelop) bool {
	if msg.Type != b.t {
		return true
	}

	b.countLock.Lock()
	b.count = b.count + 1
	count := b.count
	b.countLock.Unlock()

	if count <= b.startCount {
		return true
	} else if count <= b.endCount {
		return false
	}
	return true
}

// Filter that randomly picks a peer, block messages numbered start to end to that peer
type RandomBlockNAllow struct {
	t    string
	peer types.ReplicaID

	startCount int
	endCount   int

	count     int
	countLock *sync.Mutex
}

func NewRandomBlockNAllow(t string, start, end int) *RandomBlockNAllow {
	return &RandomBlockNAllow{
		t:          t,
		peer:       "",
		startCount: start,
		endCount:   end,
		count:      0,
		countLock:  new(sync.Mutex),
	}
}

func (b *RandomBlockNAllow) Test(msg *ControllerMsgEnvelop) bool {
	if b.peer == "" {
		b.peer = msg.To
	}

	if msg.Type != b.t {
		return true
	}

	b.countLock.Lock()
	b.count = b.count + 1
	count := b.count
	b.countLock.Unlock()

	if count <= b.startCount {
		return true
	} else if count <= b.endCount {
		return false
	}
	return true
}
