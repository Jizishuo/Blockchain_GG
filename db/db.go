package db

import (
	"Blockchain_GG/serialize/cp"
	"Blockchain_GG/utils"
)

var (
	logger   = utils.NewLogger("db")
	instance db //实例
)

type db interface {
	Init(path string) error
	HasGenesis() bool //成因
	PutGenesis(block *cp.Block) error
	PutBlock(block *cp.Block, height uint64) error
	GetHash(height uint64) ([]byte, error)

	GetHeaderViaHeight(height uint64) (*cp.BlockHeader, []byte, error) //获取标题维亚高度
	GetHeaderViaHash(h []byte) (*cp.BlockHeader, uint64, error)

	GetBlockViaHeight(height uint64) (*cp.Block, []byte, error)
	GetBlockViaHash(h []byte) (*cp.Block, uint64, error)

	GetEvidenceViaHash(h []byte) (*cp.Evidence, uint64, error)
	GetEvidenceViaKey(pubKey []byte) ([][]byte, []uint64, error)

	HasEvidence(h []byte) bool
	GetScoreViaKey(pubKey []byte) (uint64, error)
	GetLatestHeight() (uint64, error)
	GetLatestHeader() (*cp.BlockHeader, uint64, []byte, error)
	Close()
}

func Init(path string) error {
	instance = newBadger()
	return instance.Init(path)
}

func HasGenesis() bool {
	return instance.HasGenesis()
}
func PutGenesis(block *cp.Block) error {
	return instance.PutGenesis(block)
}
func PutBlock(block *cp.Block, height uint64) error {
	return instance.PutBlock(block, height)
}
func GetHash(height uint64) ([]byte, error) {
	return instance.GetHash(height)
}

func GetHeaderViaHeight(height uint64) (*cp.BlockHeader, []byte, error) {
	return instance.GetHeaderViaHeight(height)
}

func GetHeaderViaHash(height []byte) (*cp.BlockHeader, uint64, error) {
	return instance.GetHeaderViaHash(height)
}

func GetBlockViaHeight(height uint64) (*cp.Block, []byte, error) {
	return instance.GetBlockViaHeight(height)
}

func GetBlockViaHash(h []byte) (*cp.Block, uint64, error) {
	return instance.GetBlockViaHash(h)
}

func GetEvidenceViaHash(h []byte) (*cp.Evidence, uint64, error) {
	return instance.GetEvidenceViaHash(h)
}
func GetEvidenceViaKey(pubKey []byte) ([][]byte, []uint64, error) {
	return instance.GetEvidenceViaKey(pubKey)
}

func HasEvidence(h []byte) bool {
	return instance.HasEvidence(h)
}
func GetScoreViaKey(pubKey []byte) (uint64, error) {
	return instance.GetScoreViaKey(pubKey)
}
func GetLatestHeight() (uint64, error) {
	return instance.GetLatestHeight()
}
func GetLatestHeader() (*cp.BlockHeader, uint64, []byte, error) {
	return instance.GetLatestHeader()
}

func Close() {
	if instance != nil {
		instance.Close()
	}
}
