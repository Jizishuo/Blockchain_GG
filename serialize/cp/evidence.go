package cp

import (
	"Blockchain_GG/utils"
	"bytes"
	"encoding/binary"
)

const (
	EvidenceMaxDescriptionLen = 140
	EvidenceBasicLen = 1+4+1+1+1+2+2
)

// 证据
type Evidence struct {
	Version uint8
	Nonce uint32
	Hash []byte
	Description []byte
	PubKey []byte
	Sig []byte
	pc *powCache
}

func (e *Evidence) Marshal() []byte {
	result := new(bytes.Buffer)
	binary.Write(result, binary.BigEndian, e.Version)
	binary.Write(result, binary.BigEndian, e.Nonce)
	hashLen := utils.Uint8Len(e.Hash)
	binary.Write(result, binary.BigEndian, hashLen)
	binary.Write(result, binary.BigEndian, e.Hash)

	descriptionLen := utils.Uint16Len(e.Description)
	binary.Write(result, binary.BigEndian, descriptionLen)
	binary.Write(result, binary.BigEndian, e.Description)

	pubKeyLen := utils.Uint8Len(e.PubKey)
	binary.Write(result, binary.BigEndian, pubKeyLen)
	binary.Write(result, binary.BigEndian, e.PubKey)

	sigLen := utils.Uint16Len(e.Sig)
	binary.Write(result, binary.BigEndian, sigLen)
	binary.Write(result, binary.BigEndian, e.Sig)
	return result.Bytes()
}