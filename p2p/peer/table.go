package peer

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type tableImp struct {
	selfID string
	seeds map[string]*pstate // "ip:port" as key
	peers map[string]*pstate // id as key id = base32(compressPubKey)
	coolingPeers map[string]time.Time // id as key
	r *rand.Rand
	lock sync.Mutex
}


// table matains 所有peer 信息
type table interface {
	addPeers(p []*Peer, isSeed bool)
	getPeers(expect int, exclude map[string]bool) []*Peer
	exists(id string) bool

	getPeersToPing() []*Peer
	getPeersToGetNeighbours() []*Peer

	recvPing(p *Peer)
	recvPong(p *Peer)

	refresh()
}

func newTable(selfID string) table {
	return &tableImp{
		selfID:       selfID,
		seeds:        make(map[string]*pstate),
		peers:        make(map[string]*pstate),
		coolingPeers: make(map[string]time.Time),
		r:            rand.New(rand.NewSource(time.Now().Unix())),
	}
}

func (t *tableImp) addPeers(p []*Peer, isSeed bool) {
	t.lock.Lock()
	defer t.lock.Unlock()

	for _, peer := range p {
		pst := newPState(peer, isSeed)
		if isSeed {
			addr := fmt.Sprintf("%s:%d", peer.IP, peer.Port)
			t.seeds[addr] = pst
		} else {
			t.add(pst)
		}
	}
}
func (t *tableImp) getPeers(expect int, exclude map[string]bool) []*Peer {}
func (t *tableImp) exists(id string) bool {}

func (t *tableImp) getPeersToPing() []*Peer {}
func (t *tableImp) getPeersToGetNeighbours() []*Peer {}

func (t *tableImp) recvPing(p *Peer) {}
func (t *tableImp) recvPong(p *Peer) {}

func (t *tableImp) refresh() {}

func (t *tableImp) add(pst *pstate) {
	if _, ok := t.coolingPeers[pst.ID]; ok {
		return
	}
	if pst.ID == t.selfID {
		return
	}
	if _, ok := t.peers[pst.ID]; !ok {
		logger.Debug("add peer %v\n", pst)
		t.peers[pst.ID] = pst
	}
}