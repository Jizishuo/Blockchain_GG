package utils

import (
	"net"
	"time"
)

const (
	udpRecvQSize      = 1024
	udpRecvBufferSize = 1024
	udpRecvTimeout    = 2 * time.Second
	udpSendQSize      = 1024
)

type UDPPacket struct {
	Data []byte
	Addr *net.UDPAddr
}

type udpServer struct {
	ip    net.IP
	port  int
	conn  *net.UDPConn
	recvQ chan *UDPPacket
	sendQ chan *UDPPacket
	lm    *LoopMode
}

type UDPServer interface {
	GetRecvChannel() <-chan *UDPPacket
	Send(packet *UDPPacket)
	Start() bool
	Stop()
}

func NewUDPServer(ip net.IP, port int) UDPServer {
	return &udpServer{
		ip:    ip,
		port:  port,
		recvQ: make(chan *UDPPacket, udpRecvQSize),
		sendQ: make(chan *UDPPacket, udpSendQSize),
		lm:    NewLoop(2),
	}
}

// 只能接收通道
func (u *udpServer) GetRecvChannel() <-chan *UDPPacket {
	return u.recvQ
}
func (u *udpServer) Send(packet *UDPPacket) {
	select {
	case u.sendQ <- packet:
	default:
		logger.Warnln("udp server sendQ is full, drop packet")
	}
}
func (u *udpServer) Start() bool {
	udpAddr := &net.UDPAddr{
		IP:   u.ip,
		Port: u.port,
	}
	var err error
	if u.conn, err = net.ListenUDP("udp", udpAddr); err != nil {
		logger.Warn("setup UDP server failed:%v\n", err)
		return false
	}
	go u.recv()
	go u.send()
	u.lm.StartWorking()
	return true
}
func (u *udpServer) Stop() {
	if u.lm.Stop() {
		u.conn.Close()
	}
}

func (u *udpServer) recv() {
	u.lm.Add()
	defer u.lm.Done()
	for {
		select {
		case <-u.lm.D:
			return
		default:
			packBuf := make([]byte, udpRecvBufferSize) // 1024
			// 设置读操作绝对期限
			u.conn.SetReadDeadline(time.Now().Add(udpRecvTimeout))
			// 从c读取一个UDP数据包，将有效负载拷贝到packBuf，返回拷贝字节数和数据包来源地址
			n, Addr, err := u.conn.ReadFromUDP(packBuf)

			if err != nil {
				if err, ok := err.(net.Error); ok && err.Timeout() {
					break
				}
				logger.Warn("udp server read err: %v\n", err)
				break
			}

			pkt := &UDPPacket{
				Data: packBuf[:n],
				Addr: Addr,
			}
			select {
			case u.recvQ <- pkt:
			default:
				logger.Warnln("upd server recvQ is full, drop packet")
			}
		}
	}
}

func (u *udpServer) send() {
	u.lm.Add()
	defer u.lm.Done()
	for {
		select {
		case <-u.lm.D:
			return
		case packet := <-u.sendQ:
			// WriteToUDP通过c向地址addr发送一个数据包，b为包的有效负载，返回写入的字节
			_, err := u.conn.WriteToUDP(packet.Data, packet.Addr)
			if err != nil {
				logger.Warn("udp server send to %v failed: %v, sieze; %v\n",
					packet.Addr, err, len(packet.Data))
			}
		}
	}

}
