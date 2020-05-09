package discover

import (
	"Blockchain_GG/utils"
	"bytes"
	"encoding/binary"
)

type Ping struct {
	*Head
	PubKey []byte
}

func NewPing(pubKey []byte) *Ping {
	return &Ping{
		Head:NewHeadV1(MsgPing),
		PubKey: pubKey,
	}
}

func (p *Ping) Marshal() []byte {
	result := new(bytes.Buffer)
	binary.Write(result, binary.BigEndian, p.Head.Marshal())
	pubKeyLen := utils.Uint8Len(p.PubKey)
	binary.Write(result, binary.BigEndian, pubKeyLen)
	binary.Write(result, binary.BigEndian, p.PubKey)
	return result.Bytes()
}