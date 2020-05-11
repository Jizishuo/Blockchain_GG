package p2p

import (
	"Blockchain_GG/p2p/peer"
	"Blockchain_GG/utils"
)

const (
	handshakeProtocolID = 0
	nonceSize = 12
)

// 谈判
type negotiator interface {
	handshakeIo(conn utils.TCPConn, peer *peer.Peer) (codec, error)
	recvHandshake(conn utils.TCPConn, accept bool) (*peer.Peer, codec, error)
}
