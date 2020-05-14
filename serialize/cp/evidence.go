package cp

import (
	"Blockchain_GG/utils"
	"bytes"
	"encoding/binary"
	"io"
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

func NewEvidenceV1(hash, description, pubKey []byte) *Evidence {
	return &Evidence{
		Version: CoreProtocolV1,
		Nonce: 0,
		Hash: hash,
		Description: description,
		PubKey: pubKey,
		Sig: nil,
		pc: newPowCache(),
	}
}

func UnmarshalEvidence(data io.Reader) (*Evidence, error) {
	result := &Evidence{}
	var hashLen uint8
	var descriptionLen uint16
	var pubKeyLen uint8
	var sigLen uint16
	var err error

	if err = binary.Read(data, binary.BigEndian, &result.Version);err!=nil {
		return nil, err
	}
	if err = binary.Read(data, binary.BigEndian, &result.Nonce); err!=nil {
		return nil, err
	}
	if err = binary.Read(data, binary.BigEndian, &hashLen);err!=nil {
		return nil, err
	}
	result.Hash = make([]byte, hashLen)
	if err = binary.Read(data, binary.BigEndian, result.Hash);err!=nil {
		return nil, err
	}
	if err = binary.Read(data, binary.BigEndian, &descriptionLen);err!=nil{
		return nil, err
	}
	result.Description = make([]byte, descriptionLen)
	if err = binary.Read(data, binary.BigEndian,result.Description);err!=nil {
		return nil, err
	}
	if err = binary.Read(data, binary.BigEndian, &pubKeyLen); err!=nil {
		return nil, err
	}
	result.PubKey = make([]byte, pubKeyLen)
	if err = binary.Read(data, binary.BigEndian, result.PubKey);err!=nil {
		return nil, err
	}
	if err = binary.Read(data, binary.BigEndian, &sigLen);err!=nil {
		return nil, err
	}
	result.Sig = make([]byte, sigLen)
	if err = binary.Read(data, binary.BigEndian, result.Sig);err!=nil {
		return nil, err
	}
	return result, nil

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