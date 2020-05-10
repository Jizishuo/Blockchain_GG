package peer

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	coolingTime = peerExpiredTime*2
	coolingExpiredTime = 5*time.Minute
)

// 表imp(进出口)
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
		// 给定的种子创建一个伪随机资源
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
func (t *tableImp) getPeers(expect int, exclude map[string]bool) []*Peer {
	var peers []*Peer
	t.lock.Lock()
	for _, peer := range t.peers {
		if _, ok := exclude[peer.ID]; !ok && peer.isAvaible() {
			peers = append(peers, peer.Peer)
		}
	}
	t.lock.Unlock()

	peerSize := len(peers)
	if peerSize <= expect {
		return peers
	}
	for i:=0; i <peerSize; i++ {
		// 一个取值范围在[0,n)的伪随机int值，如果n<=0会panic
		j := t.r.Intn(peerSize)
		peers[i], peers[j] = peers[j], peers[i]
	}
	return peers[:expect]
}
func (t *tableImp) exists(id string) bool {
	t.lock.Lock()
	defer t.lock.Unlock()

	_, ok := t.peers[id]
	return ok
}

func (t *tableImp) getPeersToPing() []*Peer {
	t.lock.Lock()
	defer t.lock.Unlock()
	var result []*Peer
	for _, peer := range t.peers {
		if peer.isTimeToPing() {
			result = append(result, peer.Peer)
			peer.doPing()
		}
	}

	for _, seed := range t.seeds {
		result = append(result, seed.Peer)
	}
	return result

}
func (t *tableImp) getPeersToGetNeighbours() []*Peer {
	t.lock.Lock()
	defer t.lock.Unlock()

	var result []*Peer
	for _, peer := range t.peers {
		if peer.isTimeToGetNeighbours() {
			result =append(result, peer.Peer)
			peer.updataGetNeighbourTime()
		}
	}
	return result
}

func (t *tableImp) recvPing(p *Peer) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if _, ok := t.peers[p.ID]; ok {
		return
	}
	// 删除 在cool里的peer
	if _, ok := t.coolingPeers[p.ID]; ok {
		delete(t.coolingPeers, p.ID)
	}
	// 添加新的peer
	pst := newPState(p, false)
	t.add(pst)
}
func (t *tableImp) recvPong(p *Peer) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if peer, ok := t.peers[p.ID];ok {
		peer.updataActiveTime()
		return
	}
	addr := fmt.Sprintf("%s:%d", p.ID, p.Port)
	if _, ok := t.seeds[addr]; ok {
		pst := newPState(p, true)
		pst.updataActiveTime()
		t.add(pst)
		delete(t.seeds, addr)
	}
}

func (t *tableImp) refresh() {
	t.lock.Lock()
	defer t.lock.Unlock()

	for _, peer := range t.peers {
		if peer.isToRemove() {
			logger.Debug("p2p peer %v timeout, clean\n", peer.Peer)
			delete(t.peers, peer.ID)
		}
	}
	curr := time.Now()
	for k, v := range t.coolingPeers {
		if curr.Sub(v) > coolingExpiredTime {
			delete(t.coolingPeers, k)
		}
	}

}

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