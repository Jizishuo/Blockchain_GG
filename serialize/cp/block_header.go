package cp


type BlockHeader struct {
	Version uint8
	Time int64
	Nonce uint32
	Target uint32
	LastHash []byte
	Miner []byte
	EvidenceRoot []byte
	pc *powChache
}

func NewBlockHeaderV1(lastHash []byte, miner []byte, root []byte) *BlockHeader {
	return &BlockHeader{
		Version: Core
	}
}