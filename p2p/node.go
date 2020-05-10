package p2p

import (
	"Blockchain_GG/p2p/peer"
	"Blockchain_GG/params"
	"Blockchain_GG/utils"
	"github.com/btcsuite/btcd/btcec"
)

var logger = utils.NewLogger("p2p")

// Config is configs for the p2p network Node
type Config struct {
	NodeIP string
	NodePort int
	Provider peer.Provider
	MaxPeerNum int
	PrivKey *btcec.PrivateKey
	Type params.NodeType
	ChainID uint8
}
// Node is a node that can communicate with others in the p2p network.
type Node interface {
	AddProtocol(p Protocol) Pro
}