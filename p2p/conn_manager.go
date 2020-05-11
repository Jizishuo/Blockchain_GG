package p2p

import (
	"Blockchain_GG/p2p/peer"
	"Blockchain_GG/utils"
	"fmt"
	"sync"
)

// connmanager管理所有的连接
type connManager interface {
	start()
	stop()
	size() int
	getIDs() []string
	isExist(peerID string) bool
	send(p Protocol, dp *PeerData) error
	add(peer *peer.Peer, conn utils.TCPConn, ec codec, handler recvHandller) error
	String() string
}
type connManagerImp struct {
	mutex sync.Mutex
	conns map[string]*conn // <peer id, conn>
	maxPeerNum int
	removing chan string
	lm *utils.LoopMode
}

func newConnManager(maxPeerNum int) connManager {
	return &connManagerImp{
		conns: make(map[string]*conn),
		maxPeerNum: maxPeerNum,
		removing: make(chan string, maxPeerNum),
		lm: utils.NewLoop(1),
	}
}

func (c *connManagerImp) start() {
	go c.loop()
	c.lm.StartWorking()
}
func (c *connManagerImp) stop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.lm.Stop() {
		for _, conn := range c.conns {
			conn.stop()
		}
	}
}
func (c *connManagerImp) size() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.conns)
}
func (c *connManagerImp) getIDs() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var result []string
	for key := range c.conns {
		result = append(result, key)
	}
	return result
}

func (c *connManagerImp) isExist(peerID string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, ok := c.conns[peerID]
	return ok
}

func (c *connManagerImp) send(p Protocol, dp *PeerData) error {
	if c.size() == 0 {
		return ErrNoPeers
	}
	// broadcast(广播)
	if len(dp.Peer) == 0 {
		c.mutex.Lock()
		for _, conn := range c.conns {
			conn.send(p.ID(), dp.Data)
		}
		c.mutex.Unlock()
		return nil
	}
	// unicast
	c.mutex.Lock()
	conn, ok := c.conns[dp.Peer]
	c.mutex.Unlock()
	if !ok {
		return ErrPeerNotFound{Peer: dp.Peer}
	}
	conn.send(p.ID(), dp.Data)
	return nil
}
func (c *connManagerImp) add(peer *peer.Peer, conn utils.TCPConn, ec codec, handler recvHandller) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.conns[peer.ID]; ok {
		return fmt.Errorf("already exist a connetion with %s", peer.ID)
	}

	if len(c.conns) >= c.maxPeerNum {
		return fmt.Errorf("over max peer(%d) limits", len(c.conns))
	}


}
func (c *connManagerImp) String() string