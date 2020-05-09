package peer

import (
	"Blockchain_GG/crypto"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"net"
)

// // Peer is a node that can 连接 to
type Peer struct {
	IP net.IP
	Port int
	Key *btcec.PublicKey
	//我们使用 base32（压缩公钥）作为 peer ID
	//而不是像base58（哈希（公钥））等其他人员）。
	//由于哈希用于隐藏真正的 onwer
	//从未发送过任何交易的硬币，
	//和 区块链不支持交易，
	//因此，ID 只是另一个可读的表示形式公钥。

	ID string
}

// NewPeer create a Peer, key might be nil if you don't know it
func NewPeer(ip net.IP, port int, key *btcec.PublicKey) *Peer {
	p := &Peer{
		IP: ip,
		Port: port,
		Key: key,
	}
	if key != nil {
		p.ID = crypto.PubKeyToID(key)
	}
	return p
}

func (p *Peer) Address() string {
	v4IP := p.IP.To4()
	if v4IP != nil {
		return fmt.Sprintf("%s:%d", v4IP.String(), p.Port)
	}
	return fmt.Sprintf("[%s]:%d", p.IP.String(), p.Port)
}

func (p *Peer) String() string {
	return fmt.Sprintf("ID %s address %s", p.ID, p.Address())
}

type Provider interface {
	Start()
	Stop()

	// GetPeers returns avaliable peers for the caller
	GetPeers(expect int, exclude map[string]bool) ([]*Peer, error)
	// AddSeeds adds seeds for provider's initilization
	// the seeds' Peer.Key should be nil
	AddSeeds(seeds []*Peer)
}