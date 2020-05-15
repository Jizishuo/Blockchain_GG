package core

const evdsCacheSize = 1024

type RawEvidence struct {
	Hash []byte
	Description []byte  // 描述
}

type weighte