package cp

// 块
type Block struct {
	*BlockHeader
	Evds []*Evidence
}

func NewBlock(header *BlockHeader, evds []*Evidence) *Block {
	return &Block{
		BlockHeader:header,
		Evds: evds,
	}
}