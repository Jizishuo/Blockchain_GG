package discover

import (
	"bytes"
	"encoding/binary"
	"io"
	"time"
)

type Head struct {
	Version  uint8
	Type     DiscvMsgType //uint8
	Time     int64
	Reserved uint16
}

func NewHeadV1(t DiscvMsgType) *Head {
	return &Head{
		Version:  DiscoverV1,
		Type:     t,
		Time:     time.Now().Unix(),
		Reserved: uint16(0),
	}
}

func (h *Head) Marshal() []byte {
	result := new(bytes.Buffer)
	binary.Write(result, binary.BigEndian, h.Version)
	binary.Write(result, binary.BigEndian, h.Type)
	binary.Write(result, binary.BigEndian, h.Time)
	binary.Write(result, binary.BigEndian, h.Reserved)
	return result.Bytes()
}

func UnmarshalHead(data io.Reader) (*Head, error) {
	result := &Head{}
	if err := binary.Read(data, binary.BigEndian, &result.Version); err != nil {
		return nil, err
	}
	if err := binary.Read(data, binary.BigEndian, &result.Type); err != nil {
		return nil, err
	}
	if err := binary.Read(data, binary.BigEndian, &result.Time); err != nil {
		return nil, err
	}
	if err := binary.Read(data, binary.BigEndian, &result.Reserved); err != nil {
		return nil, err
	}
	return result, nil
}
