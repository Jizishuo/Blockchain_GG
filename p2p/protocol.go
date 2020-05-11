package p2p


// 协议是 p2p 网络协议必须实现的接口
// 协议
type Protocol interface {
	ID() uint8
	Name() string
}

// 协议管理程序定义用于访问 p2p 网络的接口
type ProtocolRunner interface {
	// 将数据发送到网络
	// 在成功时返回零，或 ErrPeer 未找到，ErrNoPeers 失败
	Send(dp *PeerData) error
	// GetRecvChan 返回一个获取网络数据的通道
	GetRecvChan() <- chan *PeerData
}


// PeerData 是从 ne2ks 发送或接收时使用的数据结构
type PeerData struct {
	// 对等体是发送目标或接收源节点 ID
	// 如果它是一个空字符串，则意味着广播到每个节点
	Peer string
	Data []byte
}

// ErrPeer 未找到意味着未找到对等体
type ErrPeerNotFound struct {
	Peer string
}
