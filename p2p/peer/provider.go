package peer

import (
	"Blockchain_GG/utils"
	"Blockchain_GG/crypto"
	"github.com/btcsuite/btcd/btcec"
	"net"
	"time"
)

var (
	logger = utils.NewLogger("discover")
)

type provider struct {
	ip net.IP
	port int
	compressedKey []byte
	udp utils.UDPServer
	table table
	pingHash map[string]time.Time // hash as key

	lm *utils.LoopMode
}



func NewProvider(ipstr string, port int, publicKey *btcec.PublicKey) Provider {
	ip := net.ParseIP(ipstr)
	if ip == nil {
		logger.Fatal("invalid ip: %s\n", ipstr)
	}
	p := &provider{
		ip: ip,
		port: port,
		compressedKey: publicKey.SerializeCompressed(),
		table: newTable(crypto.PubKeyToID(publicKey)),
		pingHash: make(map[string]time.Time),
		lm:utils.NewLoop(1),
	}
	p.udp = utils.NewUDPServer(ip, port)
	return p
}