package discover

import (
	"Blockchain_GG/utils"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// 笔记
type Node struct {
	Addr   *Address
	PubKey []byte
}

func NewNode(addr *Address, pubKey []byte) *Node {
	return &Node{
		Addr:   addr,
		PubKey: pubKey,
	}
}

func (n *Node) Marshal() []byte {
	result := new(bytes.Buffer)
	binary.Write(result, binary.BigEndian, n.Addr.Marshal())
	pubKeyLen := utils.Uint8Len(n.PubKey)
	binary.Write(result, binary.BigEndian, pubKeyLen)
	binary.Write(result, binary.BigEndian, n.PubKey)
	return result.Bytes()
}

func UnmarshalNode(data io.Reader) (*Node, error) {
	result := &Node{}
	var pubKeyLen uint8
	var err error
	if result.Addr, err = UnmarshalAddress(data); err != nil {
		return nil, err
	}
	if err = binary.Read(data, binary.BigEndian, &pubKeyLen); err != nil {
		return nil, err
	}
	result.PubKey = make([]byte, pubKeyLen)
	if err = binary.Read(data, binary.BigEndian, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (n *Node) String() string {
	return fmt.Sprintf("Addr %v PubKey %X", n.Addr, n.PubKey)
}
