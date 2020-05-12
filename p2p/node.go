package p2p

import (
	"Blockchain_GG/p2p/peer"
	"Blockchain_GG/params"
	"Blockchain_GG/utils"
	"fmt"
	"Blockchain_GG/crypto"
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
	AddProtocol(p Protocol) ProtocolRunner  //加协议
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
		protocols:   make(map[uint8]*protocolRunner),
		ngBlackList: make(map[string]time.Time),
		tcpConnectFunc: utils.TCPConnectTo,
		connectTask: make(chan *peer.Peer, c.MaxPeerNum),
		connMgr: newConnManager(c.MaxPeerNum),
		lm: utils.NewLoop(1),
	}
	var ip net.IP
	if ip = net.ParseIP(c.NodeIP); ip == nil {
		logger.Fatal("parse ip for tcp server failed: %s\n", c.NodeIP)
	}
	n.tcpServer = utils.NewTCPServer(ip, c.NodePort)
	return n
}

type node struct {
	tcpServer utils.TCPServer
	privKey *btcec.PrivateKey // 私钥
	chainID uint8
	nodeType params.NodeType

	maxPeersNum int
	peerProvider peer.Provider

	protocolMutex sync.Mutex
	protocols map[uint8]*protocolRunner //<Protocol ID, ProtocolRunner>

	ng negotiator // 谈判
	ngMutex sync.Mutex
	ngBlackList map[string]time.Time

	// 在测试中容易模拟
	tcpConnectFunc func(ip net.IP, port int) (utils.TCPConn, error)
	connectTask chan *peer.Peer
	connMgr connManager
	lm *utils.LoopMode
}

// 添加协议
// AddProtocol 添加运行时 p2p 网络协议
func (n *node) AddProtocol(p Protocol) ProtocolRunner {
	n.protocolMutex.Lock()
	defer n.protocolMutex.Unlock()

	if v, ok := n.protocols[p.ID()]; ok {
		logger.Fatal("protocol conflicts in ID: %s, exists: %s, wanted to add: %s",
			p.ID(), v.protocol.Name(), v.protocol.Name())
	}
	runner := newProtocolRunner(p, n.send)
	n.protocols[p.ID()] = runner
	return runner
}
func (n *node) Start() {
	if !n.tcpServer.Start() {
		logger.Fatalln("start node's tcp server failed")
	}
	n.connMgr.start()
	go n.loop()
	n.lm.StartWorking()
}
func (n *node) Stop() {
	if n.lm.Stop() {
		n.tcpServer.Stop()
		n.connMgr.stop()
	}
}

func(n *node) String() string {
	return fmt.Sprintf("[node] listen on %v\n", n.tcpServer.Addr())
}

func (n *node) loop() {
	n.lm.Add()
	defer n.lm.Done()

	// 返回一个新的Ticker，该Ticker包含一个通道字段，
	//并会每隔时间段d就向该通道发送当时的时间。
	//它会调整时间间隔或者丢弃tick信息以适应反应慢的接收者。
	//如果d<=0会panic。关闭该Ticker可以释放相关资源
	getPeersToConnectTicker := time.NewTicker(time.Second*10)
	statusReportTicker := time.NewTicker(time.Second*15)
	ngBlackListCleanTicker := time.NewTicker(time.Minute*1)

	acceptConn := n.tcpServer.GetTCPAcceptConnChannel()
	for {
		select {
		case <- n.lm.D:
			return
		case <- getPeersToConnectTicker.C:
			n.getPeersToConnect()
		case <- statusReportTicker.C:
			n.statusReport()
		case <- ngBlackListCleanTicker.C:
			n.cleanNgBlackList()
		case newPeer := <- n.connectTask:
			go func() {
				n.lm.Add()
				n.setupConn(newPeer)
				n.lm.Done()
			}()
		case newPeerConn := <- acceptConn:
			go func() {
				n.lm.Add()
				newPeerConn.SetSplitFunc(splitTCPStream)
				n.recvConn(newPeerConn)
				n.lm.Done()
			}()
		}
	}

}

func (n *node) getPeersToConnect() {
	peersNum := n.connMgr.size()
	if peersNum >= n.maxPeersNum {
		return
	}
	// 期望数
	expectNum := n.maxPeersNum - peersNum
	// 排除peers
	excludePeers := n.getExcludePeers()
	newPeers, err := n.peerProvider.GetPeers(expectNum, excludePeers)
	if err != nil {
		logger.Warn("get peers from provider failed:%v\n", err)
		return
	}

	for _, newPeer := range newPeers {
		n.connectTask <- newPeer
	}
}

func (n *node) statusReport() {
	if utils.GetLogLevel() < utils.LogDebugLevel {
		return
	}
	logger.Debug("current address book: %v\n", n.connMgr)
}

func (n *node) setupConn(newPeer *peer.Peer) {
	// 始终假设远程站点将同时建立连接;
	// 比较 ID，较小的 ID 将是客户端
	if crypto.PrivKeyToID(n.privKey) > newPeer.ID {
		time.Sleep(time.Second*10)
	}
	if n.connMgr.isExist(newPeer.ID) {
		return
	}

	conn, err := n.tcpConnectFunc(newPeer.IP, newPeer.Port)
	if err != nil {
		logger.Warn("setup connection to %v failed %v", newPeer, err)
		return
	}
	conn.SetSplitFunc(splitTCPStream)
	ec, err := n.ng.handshakeIo(conn, newPeer)
	if err != nil {
		logger.Warn("handshake to %v failed: %v", newPeer, err)
		conn.Disconnect()
		n.addNgBlackList(newPeer.ID)
		return
	}
	n.addConn(newPeer, conn, ec)
}

func (n* node) recvConn(conn utils.TCPConn) {
	accept := false
	if n .connMgr.size() < n.maxPeersNum {
		accept = true
	}
	peer ,ec, err := n.ng.recvHandshake(conn, accept)
	if err != nil {
		logger.Warn("handle handshake from remote failed: %v\n", err)
		conn.Disconnect()
		return
	}
	if !accept {
		conn.Disconnect()
		return
	}
	n.addConn(peer, conn, ec)
}

func(n *node) addConn(peer *peer.Peer, conn utils.TCPConn, ec codec) {
	if err := n.connMgr.add(peer, conn, ec, n.recv); err != nil {
		logger.Debug("addConn failed :%v\n", err)
		conn.Disconnect()
	}
}

func (n *node) send(p Protocol, dp *PeerData) error {
	return n.connMgr.send(p, dp)
}

func (n *node) recv(peer string, protocolID uint8, data []byte) {
	if runner, ok := n.protocols[protocolID]; ok {
		select {
		case runner.Data <- &PeerData{
			Peer: peer,
			Data: data,
		}:
		default:
			logger.Warn("protocol %s recv packet queue full, drop it",
				runner.protocol.Name())
		}
	}
}

func (n *node) addNgBlackList(peerID string) {
	n.ngMutex.Lock()
	defer n.ngMutex.Unlock()
	n.ngBlackList[peerID] = time.Now()
}

func (n *node) cleanNgBlackList() {
	n.ngMutex.Lock()
	defer n.ngMutex.Unlock()

	curr := time.Now()
	for k, v := range n.ngBlackList {
		if curr.Sub(v) > time.Minute*30 {
			delete(n.ngBlackList, k)
		}
	}
}

func (n *node) getExcludePeers() map[string]bool {
	result := make(map[string]bool)

	n.ngMutex.Lock()
	for k := range n.ngBlackList {
		result[k] = true
	}
	n.ngMutex.Unlock()

	connectedID := n.connMgr.getIDs()
	for _, id := range connectedID {
		result[id] = true
	}
	return result
}