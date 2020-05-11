package p2p

import (
	"Blockchain_GG/p2p/peer"
	"Blockchain_GG/params"
	"Blockchain_GG/utils"
	"github.com/btcsuite/btcd/btcec"
	"net"
	"sync"
	"time"
)

var logger = utils.NewLogger("p2p")

// 配置是 p2p 网络节点的配置
type Config struct {
	NodeIP string
	NodePort int
	Provider peer.Provider
	MaxPeerNum int
	PrivKey *btcec.PrivateKey
	Type params.NodeType
	ChainID uint8
}
// 节点是一个节点，可以与其他人在p2p网络中通信。
type Node interface {
	AddProtocol(p Protocol) ProtocolRunner
	Start()
	Stop()
}

// NewNode returns a p2p network Node
func NewNode(c *Config) Node {
	if c.Type != params.FullNode && c.Type != params.LightNode {
		logger.Fatal("invalid node type &d\n", c.Type)
	}
	n := &node{
		privKey: c.PrivKey,
		chainID: c.ChainID,
		nodeType: c.Type,
		maxPeersNum: c.MaxPeerNum,
		peerProvider: c.Provider,
		protocols: make(map[uint8]*ProtocolRunner),
		ngBlackList: make(map[string]time.Time),
		tcpConnectFunc: utils.TCPConnectTo,
		connectTask: make(chan *peer.Peer, c.MaxPeerNum),
		connMgr: new
	}
}

type node struct {
	tcpServer utils.TCPServer
	privKey *btcec.PrivateKey // 私钥
	chainID uint8
	nodeType params.NodeType

	maxPeersNum int
	peerProvider peer.Provider

	protocolMutex sync.Mutex
	protocols map[uint8]*ProtocolRunner //<Protocol ID, ProtocolRunner>

	ng negotiator // 谈判
	ngMutex sync.Mutex
	ngBlackList map[string]time.Time

	// 在测试中容易模拟
	tcpConnectFunc func(ip net.IP, port int) (utils.TCPConn, error)
	connectTask chan *peer.Peer
	connMgr connManager
	lm *utils.LoopMode
}