package core

import (
	"Blockchain_GG/p2p"
	"Blockchain_GG/params"
	"Blockchain_GG/utils"
	"Blockchain_GG/core/blockchain"
	"github.com/btcsuite/btcd/btcec"
)

var (
	logger = utils.NewLogger("core")
)

type Config struct {
	Node p2p.Node
	NodeType params.NodeType
	PrivKey *btcec.PrivateKey
	ParalleMine int
}

type Core struct {
	chain *blockchain.Chain
	evPool *ev
}