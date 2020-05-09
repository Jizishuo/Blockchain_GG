package peer

import (
	"Blockchain_GG/serialize/discover"
	"Blockchain_GG/utils"
	"Blockchain_GG/crypto"
	"github.com/btcsuite/btcd/btcec"
	"net"
	"time"
)

const (
	msgDiscardTime int64 = 8 // 8s
	maxNeighboursRspNum = 8  // 最大邻居数量
	pingHashExpiredTime = peerExpiredTime
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
		logger.Fatal("invalid ip:%s\n", ipstr)
	}

	p := &provider{
		ip:            ip,
		port:          port,
		compressedKey: publicKey.SerializeCompressed(),
		table:         newTable(crypto.PubKeyToID(publicKey)),
		pingHash:      make(map[string]time.Time),
		lm:            utils.NewLoop(1),
	}
	p.udp = utils.NewUDPServer(ip, port)

	return p
}

func (p *provider) Start() {
	if !p.udp.Start() {
		logger.Fatalln("start udp server failed")
	}
	go p.loop()
}

func (p *provider) Stop() {

}

func (p *provider) AddSeeds() {

}

func (p *provider) GetPeers() {

}

func (p *provider) loop() {
	p.lm.Add()
	defer p.lm.Done()
	// 返回一个新的Ticker，该Ticker包含一个通道字段，
	//并会每隔时间段d就向该通道发送当时的时间。
	//它会调整时间间隔或者丢弃tick信息以适应反应慢的接收者。
	//如果d<=0会panic。关闭该Ticker可以释放相关资源
	refrsshTicker := time.NewTicker(peerExpiredTime*2)
	taskTicker := time.NewTicker(time.Second*2)
	recvQ := p.udp.GetRecvChannel()

	for {
		select {
		case <-p.lm.D:
			return
			case <- taskTicker.C:
				p.ping()
				p.getNeighbours()
		case pkt := <- recvQ :
			p.h

		}
	}

}


func (p *provider) ping() {
	targets := p.table.getPeersToPing()
	for _, peer := range targets {
		pkt := discover.NewPing(p.compressedKey).Marshal()
		if addr, err := net.ResolveUDPAddr("udp", peer.Address()); err == nil {
			p.send(pkt, addr)
			p.pingHash[utils.ToHex(utils.Hash(pkt))] = time.Now()
		}
	}
}

func (p *provider) send(msg []byte, addr *net.UDPAddr) {
	pkt := &utils.UDPPacket{
		Data: msg,
		Addr: addr,
	}
	p.udp.Send(pkt)
}

func (p *provider) getNeighbours() {
	targets := p.table.getPeersToGetNeighbours()
	for _, peer := range targets {
		pkt := discover.NewGetNeighbours(p.compressedKey).Marshal()
		if addr, err := net.ResolveUDPAddr("udp", peer.Address()); err == nil {
			p.send(pkt, addr)
		}
	}
}

func (p *provider) handleRecv(pkt *utils.UDPPacket) {
	head, err := discover.UnmarshalHead()
}