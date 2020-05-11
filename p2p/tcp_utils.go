package p2p

import (
	"Blockchain_GG/utils"
	"bytes"
	"encoding/binary"
	"hash/crc32"
)

/*
+-------------+-----------+--------------+
|   Length    |    CRC    |    Protocol  |
+-------------+-----------+--------------+
|                Payload                 |
+----------------------------------------+

(bytes)
Length		4
CRC			4
Protocol	1
*/

const tcpHeaderSize = 9

func verifyTCPPacket(packet []byte) (bool, []byte, uint8) {
	var length uint32
	var crc uint32
	var protocolID uint8

	packetReader := bytes.NewBuffer(packet)
	binary.Read(packetReader, binary.BigEndian, &length)
	binary.Read(packetReader, binary.BigEndian, &crc)
	binary.Read(packetReader, binary.BigEndian, &protocolID)

	payload := make([]byte, length)
	packetReader.Read(payload)

	// 数据data使用IEEE多项式计算出的CRC-32校验和
	checkCrc := crc32.ChecksumIEEE(payload)
	if crc != checkCrc {
		return false, nil, 0
	}
	return true, payload, protocolID
}

func buildTCPPacket(payload []byte, protocolID uint8) []byte {
	length := utils.Uint32Len(payload)
	crc := crc32.ChecksumIEEE(payload)

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, length)
	binary.Write(buf, binary.BigEndian, crc)
	binary.Write(buf, binary.BigEndian, protocolID)
	buf.Write(payload)
	return buf.Bytes()
}