package peer

import (
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