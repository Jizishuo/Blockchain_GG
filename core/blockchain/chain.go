package blockchain

import (
	"Blockchain_GG/db"
	"Blockchain_GG/serialize/cp"
	"Blockchain_GG/utils"
	"bytes"
	"fmt"
	"sync"
	"time"
)

const (
	syncMaxBlocks uint64 = 128
	// 	alpha 是用于操作链的高度差
	//	1.如果branch_a"alpha"高于branch_b，则从缓存中删除branch_b
	//	2.如果块转发Num（向前参考块编号）为 1，
	//	它比最长的分支低"alpha"，然后将其保存到 db 并尝试将其从缓存中删除
	//	3.如果接收的块低于分支的"alpha"，则不接受

	alpha = 8
)

var logger = utils.NewLogger("chain")

type Chain struct {
	PassiveChangeNotify chan bool // 被动更改通知

	oldestBlock   *block
	branches      []*branch //分支
	longestBranch *branch   //
	lastHeight    uint64
	branchLock    sync.Mutex
	pendingBlocks chan []*cp.Block // 等待区块
	lm            *utils.LoopMode
}

// NewChain 返回链，应只调用一次
func NewChain() *Chain {
	return &Chain{
		PassiveChangeNotify: make(chan bool, 1),
		pendingBlocks:       make(chan []*cp.Block, 16),
		lm:                  utils.NewLoop(1),
	}
}

type Config struct {
	BlockTargetLimit    uint32
	EvidenceTargetLimit uint32
	BlockInterval       int
	Genesis             string
}

// 从 db 初始化链，应只调用一次
func (c *Chain) Init(conf *Config) error {
	initMiningParams(conf)

	if !db.HasGenesis() {
		logger.Info("chain starts with empty database")
		if err := c.initGenesis(conf.Genesis); err != nil {
			logger.Warn("chain init failed:%v\n", err)
			return err
		}
		return nil
	}

	return c.initFromDB()
}

func (c *Chain) Start() {
	go c.loop()
	c.lm.StartWorking()
}

func (c *Chain) Stop() {
	c.lm.Stop()
}

// 添加块将新块追加到链中
func (c *Chain) AddBlocks(blocks []*cp.Block, local bool) {
	if local {
		c.addBlocks(blocks, local)
		return
	}
	c.pendingBlocks <- blocks
}

// 下一块目标返回下一个块所需的目标
func (c *Chain) NextBlockTarget(newBlockTime int64) uint32 {
	c.branchLock.Lock()
	defer c.branchLock.Unlock()

	return c.longestBranch.nextBlockTarget(newBlockTime)
}

// 最新BlockHash返回最长的分支最新块哈希
func (c *Chain) LatestBlockHash() []byte {
	c.branchLock.Lock()
	defer c.branchLock.Unlock()
	return c.longestBranch.hash()
}

// GetSyncHash 返回同步使用的块哈希和高度差异
func (c *Chain) GetSyncHash(base []byte) (end []byte, heightDiff uint32, err error) {
	c.branchLock.Lock()
	defer c.branchLock.Unlock()

	var hdiff uint32

	// search in the longest branch
	if baseBlock := c.longestBranch.getBlock(base); baseBlock != nil {
		b := c.longestBranch.head
		if bytes.Equal(b.hash, base) {
			return nil, 0, ErrAlreadyUpToDate{base}
		}

		endHash := b.hash
		for {
			if b == nil {
				// flushing cache to db happens during this time
				return nil, 0, ErrFlushingCache{base}
			}
			if bytes.Equal(b.hash, base) {
				break
			}
			hdiff++
			b = b.backward
		}

		return endHash, hdiff, nil
	}

	// search in the db 在 db 中搜索
	_, baseHeight, err := db.GetHeaderViaHash(base)
	if err != nil {
		return nil, 0, ErrHashNotFound{base}
	}

	_, dbLatestHeight, dbLatestHash, err := db.GetLatestHeader()
	if err != nil {
		return nil, 0, err
	}

	if dbLatestHeight-baseHeight >= syncMaxBlocks {
		respHash, _ := db.GetHash(baseHeight + syncMaxBlocks)
		return respHash, uint32(syncMaxBlocks), nil
	}

	hdiff = uint32(dbLatestHeight - baseHeight)
	return dbLatestHash, hdiff, nil
}

//GetSyncBlocks 返回同步使用的块
func (c *Chain) GetSyncBlocks(base []byte, end []byte, onlyHeader bool) ([]*cp.Block, error) {
	c.branchLock.Lock()
	defer c.branchLock.Unlock()

	var result []*cp.Block

	// search in the longest branch 在最长的分支中搜索
	baseBlock := c.longestBranch.getBlock(base)
	endBlock := c.longestBranch.getBlock(end)
	if baseBlock != nil && endBlock != nil && baseBlock.height < endBlock.height {
		iter := endBlock
		for {
			if iter.height == baseBlock.height {
				// ignore the base block 忽略基块
				break
			}

			if iter == nil {
				// flushing cache to db happens during this time
				return nil, ErrFlushingCache{base}
			}

			result = append([]*cp.Block{iter.Block.ShallowCopy(onlyHeader)}, result...)
			iter = iter.backward
		}
		return result, nil
	}

	if baseBlock == nil {
		logger.Debug("cache not found base, search in db\n")
	} else if endBlock == nil {
		logger.Debug("cache not found end, search in db\n")
	} else {
		return nil, ErrInvalidBlockRange{fmt.Sprintf("block heigh error, base %d, end %d\n",
			baseBlock.height, endBlock.height)}
	}

	// search in the db
	sBaseBlock, baseHeight, _ := db.GetBlockViaHash(base)
	sEndBlock, endHeight, _ := db.GetBlockViaHash(end)
	if sBaseBlock != nil && sEndBlock != nil && baseHeight < endHeight {
		for i := baseHeight + 1; i <= endHeight; i++ {
			sBlock, _, _ := db.GetBlockViaHeight(i)
			result = append(result, sBlock.ShallowCopy(onlyHeader))
		}
		return result, nil
	}

	if sBaseBlock != nil && endBlock != nil {
		return result, ErrFlushingCache{base}
	}

	return nil, ErrHashNotFound{base}
}

// GetSyncBlockHash 返回每个分支的最新块哈希
func (c *Chain) GetSyncBlockHash() [][]byte {
	c.branchLock.Lock()
	defer c.branchLock.Unlock()

	var result [][]byte
	for _, bc := range c.branches {
		result = append(result, bc.hash())
	}
	return result
}

// 验证证据通过匹配的分支验证证据
func (c *Chain) VerifyEvidence(e *cp.Evidence) error {
	c.branchLock.Lock()
	defer c.branchLock.Unlock()

	if err := c.longestBranch.verifyEvidence(e); err != nil {
		return err
	}

	return nil
}

//获取未存储的块返回具有高度的未存储块
//结果按高度按递减顺序排序
func (c *Chain) GetUnstoredBlocks() ([]*cp.Block, []uint64) {
	c.branchLock.Lock()
	defer c.branchLock.Unlock()

	var blocks []*cp.Block
	var heights []uint64
	iter := c.longestBranch.head
	for {
		if iter == nil {
			break
		}

		if iter.isStored() {
			break
		}

		blocks = append(blocks, iter.Block)
		heights = append(heights, iter.height)
		iter = iter.backward
	}

	return blocks, heights
}

func (c *Chain) initGenesis(genesis string) error {
	var genesisB []byte
	var cb *cp.Block
	var err error

	if genesisB, err = utils.FromHex(genesis); err != nil {
		return err
	}

	if cb, err = cp.UnmarshalBlock(bytes.NewReader(genesisB)); err != nil {
		return err
	}

	if err = db.PutGenesis(cb); err != nil {
		return err
	}

	// the genesis block height is 1 成因块高度为 1
	c.initFirstBranch(newBlock(cb, 1, true))
	return nil
}

func (c *Chain) initFromDB() error {
	var beginHeight uint64 = 1
	lastHeight, err := db.GetLatestHeight()
	if err != nil {
		logger.Warn("get latest height failed:%v\n", err)
		return err
	}
	if lastHeight > ReferenceBlocks {
		// 只将最后的"参考块"块放入缓存中
		beginHeight = lastHeight - ReferenceBlocks // only takes the last 'ReferenceBlocks' blocks into cache
	}

	var blocks []*block
	for height := beginHeight; height <= lastHeight; height++ {
		cb, _, err := db.GetBlockViaHeight(height)
		if err != nil {
			return fmt.Errorf("height %d, broken db data for block", height)
		}

		blocks = append(blocks, newBlock(cb, height, true))
	}

	bc := c.initFirstBranch(blocks[0])
	for i := 1; i < len(blocks); i++ {
		bc.add(blocks[i])
	}
	return nil
}

func (c *Chain) initFirstBranch(b *block) *branch {
	bc := newBranch(b)
	c.oldestBlock = b
	c.branches = append(c.branches, bc)
	c.longestBranch = bc
	c.lastHeight = c.longestBranch.height()
	return bc
}

func (c *Chain) loop() {
	c.lm.Add()
	defer c.lm.Done()

	maintainTicker := time.NewTicker(time.Duration(2) * BlockInterval)
	statusReportTicker := time.NewTicker(BlockInterval / 2)
	for {
		select {
		case <-c.lm.D:
			return
		case <-maintainTicker.C:
			c.maintain()
		case blocks := <-c.pendingBlocks:
			c.addBlocks(blocks, false)
		case <-statusReportTicker.C:
			c.statusReport()
		}
	}
}

// maintain cleans up the chain and flush cache into db 维护清理链并将缓存刷新到 db
func (c *Chain) maintain() {
	c.branchLock.Lock()
	defer c.branchLock.Unlock()

	var reservedBranches []*branch
	for _, bc := range c.branches {
		if c.longestBranch.height()-bc.height() > alpha {
			logger.Debug("remove branch %s\n", bc.String())
			bc.remove()
			continue
		}
		reservedBranches = append(reservedBranches, bc)
	}
	c.branches = reservedBranches

	iter := c.oldestBlock
	for {
		if iter.forwardNum() != 1 {
			break
		}

		// no fork from this block 此块没有分叉
		if c.longestBranch.height()-iter.height > alpha {
			removingBlock := iter
			if !removingBlock.isStored() {
				if err := db.PutBlock(removingBlock.Block, removingBlock.height); err != nil {
					logger.Fatal("store block failed:%v\n", err)
				}
				removingBlock.stored = true
				logger.Debug("store block (height %d)\n", iter.height)
			}

			// iter++
			removingBlock.fordward.Range(func(k, v interface{}) bool {
				vBlock := v.(*block)
				iter = vBlock
				return true
			})

			// don't remove after storing immediately,
			// keep some blocks both exist in cache and db, for conveniently synchronizing
			// 			在立即存储后不要删除，
			//			保留缓存和 db 中存在的某些块，以便方便地同步
			if c.longestBranch.height()-iter.height > syncMaxBlocks {

				// disconnect removingBlock with iter
				removingBlock.removeForward(iter)
				iter.removeBackward()

				// remove in each branch's cache
				for _, bc := range c.branches {
					bc.removeFromCache(removingBlock)
				}

				c.oldestBlock = iter
			}

		} else {
			break
		}
	}
}

func (c *Chain) addBlocks(blocks []*cp.Block, local bool) {
	c.branchLock.Lock()
	defer c.branchLock.Unlock()

	if len(blocks) == 0 {
		logger.Warnln("add blocks failed:empyth blocks")
		return
	}

	var err error
	var bc *branch
	lastHash := blocks[0].LastHash
	bc = c.getBranch(lastHash)
	if bc == nil {
		if bc, err = c.createBranch(blocks[0]); err != nil {
			logger.Info("add blocks failed:%v\n", err)
			return
		}
	}

	for _, cb := range blocks {
		if err := bc.verifyBlock(cb); err != nil {
			logger.Warn("verify blocks failed:%v\n", err)
			return
		}
		bc.add(newBlock(cb, bc.height()+1, false))
	}

	if !local {
		c.notifyCheck()
	}
}

func (c *Chain) getBranch(blochHash []byte) *branch {
	for _, b := range c.branches {
		if bytes.Equal(b.hash(), blochHash) {
			return b
		}
	}
	return nil
}

//创建分止
func (c *Chain) createBranch(newBlock *cp.Block) (*branch, error) {
	var result *branch
	lastHash := newBlock.LastHash

	for _, b := range c.branches {
		if matchBlock := b.getBlock(lastHash); matchBlock != nil {
			if b.height()-matchBlock.height > alpha {
				return nil, fmt.Errorf("the block is too old, branch height %d, block height %d",
					b.height(), matchBlock.height)
			}

			if matchBlock.isBackwardOf(newBlock) {
				return nil, fmt.Errorf("duplicated new block")
			}

			logger.Info("branch fork happen at block %s height %d\n",
				utils.ToHex(matchBlock.hash), matchBlock.height)

			result = newBranch(matchBlock)
			c.branches = append(c.branches, result)
			return result, nil
		}
	}

	return nil, fmt.Errorf("not found branch for last hash %X", lastHash)
}

// 获取最长分支
func (c *Chain) getLongestBranch() *branch {
	var longestBranch *branch
	var height uint64
	for _, b := range c.branches {
		if b.height() > height {
			longestBranch = b
			height = b.height()
		} else if b.height() == height {
			//pick the random one
			if time.Now().Unix()%2 == 0 {
				longestBranch = b
			}
		}
	}
	return longestBranch
}

//通知检查
func (c *Chain) notifyCheck() {
	longestBranch := c.getLongestBranch()
	if longestBranch.height() > c.lastHeight {
		c.longestBranch = longestBranch
		c.lastHeight = c.longestBranch.height()

		select {
		case c.PassiveChangeNotify <- true:
		default:
		}
	}
}

func (c *Chain) statusReport() {
	if utils.GetLogLevel() < utils.LogDebugLevel {
		return
	}

	c.branchLock.Lock()
	defer c.branchLock.Unlock()

	branchNum := len(c.branches)
	text := "\n\toldest: %X with height %d \n\tlongest head:%X \n\tbranch number:%d, details:\n%s"

	var details string
	for i := 0; i < branchNum; i++ {
		details += c.branches[i].String() + "\n\n"
	}

	logger.Debug(text, c.oldestBlock.hash[utils.HashLength-2:], c.oldestBlock.height,
		c.longestBranch.hash()[utils.HashLength-2:], branchNum, details)
}
