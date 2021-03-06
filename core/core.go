package core

import (
	"fmt"
	"sort"

	"Blockchain_GG/core/blockchain"
	"Blockchain_GG/p2p"
	"Blockchain_GG/params"
	"Blockchain_GG/serialize/cp"
	"Blockchain_GG/utils"
	"github.com/btcsuite/btcd/btcec"
)

var (
	logger = utils.NewLogger("core")
)

type Config struct {
	Node         p2p.Node
	NodeType     params.NodeType
	PrivKey      *btcec.PrivateKey
	ParallelMine int

	*blockchain.Config
}

type Core struct {
	chain      *blockchain.Chain
	evPool     *evidencePool
	n          *net
	queryCache *qCache
	s          *scheduler
	mining     bool
}

func NewCore(conf *Config) *Core {
	chain := blockchain.NewChain()
	if err := chain.Init(conf.Config); err != nil {
		logger.Fatal("init core module failed:%v\n", err)
	}
	chain.Start()

	evPool := newEvidencePool(conf.PrivKey)

	n := newNet(conf.Node, chain, evPool, conf.NodeType)
	evPool.setBroadcastChan(n.evdsToBroadcast)
	n.start()
	evPool.start()

	queryCache := newQCache(chain)

	var s *scheduler
	mining := false
	if conf.ParallelMine == 0 {
		logger.Info("parallel mining thread number is 0, the program won't do any mining")
	} else {
		pubKey := conf.PrivKey.PubKey()
		s = newScheduler(evPool, chain, n, pubKey.SerializeCompressed(), conf.ParallelMine)
		s.start()
		mining = true
	}

	return &Core{
		chain:      chain,
		evPool:     evPool,
		n:          n,
		queryCache: queryCache,
		s:          s,
		mining:     mining,
	}
}

// Stop stops the core module working 停止停止核心模块工作
func (c *Core) Stop() {
	if c.mining {
		c.s.stop()
	}

	c.evPool.stop()
	c.n.stop()
	c.chain.Stop()
}

// UploadEvidenceRaw uploads the hash of evidence
// the node will sign it and broadcast to the network
// 上传证据原始上传证据哈希
// 节点将签名并广播到网络
func (c *Core) UploadEvidenceRaw(evds []*RawEvidence) error {
	for _, evd := range evds {
		if len(evd.Hash) != utils.HashLength {
			return fmt.Errorf("invalid hash size[%X]", evd.Hash)
		}
	}

	c.evPool.addRawEvidence(evds)
	return nil
}

// UploadEvidence uploads the evidence 上传证据上传证据
func (c *Core) UploadEvidence(evds []*cp.Evidence) error {
	for _, evd := range evds {
		if err := c.chain.VerifyEvidence(evd); err != nil {
			if _, ok := err.(blockchain.ErrEvidenceAlreadyExist); ok {
				return err
			}
			return fmt.Errorf("verify evidence failed[%v]", evd)
		}
	}

	c.evPool.addEvidence(evds, false)
	return nil
}

// 查询证据
func (c *Core) QueryEvidence(hash []string) []*EvidenceInfo {
	return c.queryCache.getEvidence(hash)
}

func (c *Core) QueryAccount(id string) ([][]byte, uint64) {
	return c.queryCache.getAccount(id)
}

func (c *Core) QueryBlockViaHeights(heights []uint64) []*BlockInfo {
	sort.Slice(heights, func(i, j int) bool {
		return heights[i] > heights[j] // from heigher to lower
	})

	var result []*BlockInfo
	for _, height := range heights {
		info := c.queryCache.getBlockViaHeight(height)
		if info != nil {
			result = append(result, info)
		}
	}

	return result
}

func (c *Core) QueryLatestBlock() *BlockInfo {
	return c.queryCache.getLatestBlock()
}

func (c *Core) QueryBlockViaRange(begin, end uint64) []*BlockInfo {
	var result []*BlockInfo
	for i := end; i >= begin; i-- {
		info := c.queryCache.getBlockViaHeight(i)
		if info != nil {
			result = append(result, info)
		}
	}

	return result
}

func (c *Core) QueryBlockViaHash(hash []string) []*BlockInfo {
	var result []*BlockInfo
	for _, h := range hash {
		info := c.queryCache.getBlockViaHash(h)
		if info != nil {
			result = append(result, info)
		}
	}

	return result
}
