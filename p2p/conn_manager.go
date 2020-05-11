package p2p

import (
	"Blockchain_GG/p2p/peer"
	"Blockchain_GG/utils"
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
func (c *connManagerImp) size() int
func (c *connManagerImp) getIDs() []string
func (c *connManagerImp) isExist(peerID string) bool
func (c *connManagerImp) send(p Protocol, dp *PeerData) error
func (c *connManagerImp) add(peer *peer.Peer, conn utils.TCPConn, ec codec, handler recvHandller) error
func (c *connManagerImp) String() string