package db

import (
	"Blockchain_GG/serialize/storage"
	"Blockchain_GG/utils"
	"Blockchain_GG/serialize/cp"
	"github.com/dgraph-io/badger"
	"path/filepath"
	"time"
)

// 占位符
var placeHolder = []byte("0")

type badgerDB struct {
	*badger.DB
	lm *utils.LoopMode
}

func newBadger() *badgerDB {
	return &badgerDB{
		lm: utils.NewLoop(1),
	}
}

func(b *badgerDB) Init(path string) error {
	var dbpath string
	var err error

	if dbpath, err = filepath.Abs(path); err != nil {
		return err
	}
	if err = utils.AccessCheck(dbpath); err != nil {
		return err
	}
	opts := badger.DefaultOptions(dbpath)
	opts = opts.WithLogger(nil)
	opts = opts.WithValueLogFileSize(512<<20)
	opts = opts.WithMaxTableSize(32<<20)

	b.DB, err = badger.Open(opts)
	if err != nil {
		return b.warpError(err)
	}
	b.start()
	return nil
}
func (b *badgerDB) Close() {
	b.stop()
	b.DB.Close()
}

func (b *badgerDB) HasGenesis() bool {
	rf := func(tx *badger.Txn) error {
		_, err := tx.Get(mGenesis)
		return err
	}
	err :=b.View(rf)
	if err == nil {
		return true
	} else if err == badger.ErrKeyNotFound {
		return false
	} else {
		logger.Fatal("check genesis failed: %v\n", err)
		return false
	}
}
func (b *badgerDB) PutGenesis(block *cp.Block) error {
	wf := func(tx *badger.Txn) error {
		if err := tx.Set(mGenesis, placeHolder); err != nil {
			return err
		}
		if err := b.putBlockTX(block, 1, tx); err != nil {
			return err
		}
		if err := b.updateLatestHeightTX(1, tx); err != nil {
			return err
		}
		return nil
	}
	return b.update(wf)
}
// PutBlock 存储块进入 db
// 它不应修改现有的块，新块的高度应增加一个
func (b *badgerDB) PutBlock(block *cp.Block, height uint64) error {
	latesHeight, err :=b.GetLatestHeight()
	if err != nil {
		return err
	}
	// 期望高度
	expectHeight := latesHeight + 1
	if height != expectHeight {
		return ErrInvalidHeight{height, expectHeight}
	}
	wf := func(tx *badger.Txn) error {
		if err := b.PutBlockTX(block, height); err != nil {
			return err
		}
		if err := b.updateLatestHeightTX(height, tx); err != nil {
			return err
		}
		return nil
	}
	return b.Update(wf)
}
// GetHash 通过其高度获得块哈希
func (b *badgerDB) GetHash(height uint64) ([]byte, error) {
	var result []byte
	hashKey := getHashKey(height)
	rf := func(tx *badger.Txn) error {
		item, err := tx.Get(hashKey)
		if err != nil {
			return err
		}
		result, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	}
	return result, b.View(rf)
}

func (b *badgerDB) GetHeaderViaHeight(height uint64) (*cp.BlockHeader, []byte, error) {
	hash, err := b.GetHash(height)
	if err != nil {
		return nil, nil, err
	}
	header, err := b.getHeader(height, hash)
	if err != nil {
		return nil, nil, err
	}
	return header.BlockHeader, hash, nil
}

func (b *badgerDB) GetHeaderViaHash(h []byte) (*cp.BlockHeader, uint64, error) {
	height, err := b.getHaderHeight(h)
	if err != nil {
		return nil, 0, err
	}
	header, err := b.getHeader(height, h)
	if err != nil {
		return nil, 0, err
	}
	return header.BlockHeader, height, nil
}

func (b *badgerDB) GetBlockViaHeight(height uint64) (*cp.Block, []byte, error) {
	hash, err := b.GetHash(height)
	if err != nil {
		return nil, nil, err
	}
	result, err := b.getCPBlock(height, hash)
	if err != nil {
		return nil, nil, err
	}
	return result, hash, err
}

func (b *badgerDB) GetBlcokViaHash(h []byte) (*cp.Block, uint64, error) {
	height, err := b.
}

func (b *badgerDB) GetEvidenceViaHash(h []byte) (*cp.Evidence, uint64, error) {
	return instance.GetEvidenceViaHash(h)
}
func (b *badgerDB) GetEvidenceViaKey(pubKey []byte) ([][]byte, []uint64, error) {
	return instance.GetEvidenceViaKey(pubKey)
}

func (b *badgerDB) HasEvidence(h []byte) bool {
	return instance.HasEvidence(h)
}
func (b *badgerDB) GetScoreViaKey(pubKey []byte) (uint64, error) {
	return instance.GetScoreViaKey(pubKey)
}
func (b *badgerDB) GetLatestHeight() (uint64, error) {
	return instance.GetLatestHeight()
}
func (b *badgerDB) GetLatesHeader() (*cp.BlockHeader, uint64, []byte, error) {
	return instance.GetLatesHeader()
}

func (b *badgerDB) getHeaderHeight(hash []byte) (uint64, error) {
	var result uint64
	headerHeightKey := getHeaderHeightKey(hash)
	rf := func(tx *badger.Txn) error {
		item, err := tx.Get(headerHeightKey)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			result  = byteh(val)
			return nil
		})
	}
	return result, b.view(rf)
}

func (b *badgerDB) putBlockTX(block *cp.Block, height uint64, tx *badger.Txn) error {
	hash := block.G
}

func (b *badgerDB) putEvidenceTX(hash []byte, evds []*cp.Evidence, height uint64, tx *badger.Txn) error {
	var evdsHash [][]byte
	for _, e := range evds {
		storageData := e.Marshal()
		if err := tx.Set(getEvidenceKey(height, e.Hash), storageData); err != nil {
			return err
		}
		if err := tx.Set(getEvidenceHeightKey(e.Hash), hbyte(height)); err != nil {
			return err
		}
		if err := b.updateAccountEvidenceTX(e, height, tx); err != nil {
			return err
		}
		evdsHash = append(evdsHash, e.Hash)
	}
	block := storage.NewBlock(evdsHash)
	storageData := block.Marshal()
	if err := tx.Set(getBlockKey(height, hash), storageData); err != nil {
		return err
	}
	return nil
}


func (b *badgerDB)updateAccountEvidenceTX(evd *cp.Evidence, height uint64, tx *badger.Txn) error {
	accountEvidenceKey := append(getAccountEvidenceKeyPrefix(evd.PubKey), evd.Hash...)
	heightValue := hbyte(height)
	return tx.Set(accountEvidenceKey, heightValue)
}

func (b *badgerDB) updateLatesHeightTX(height uint64, tx *badger.Txn) error {
	if err := tx.Set(mLatestHeight, hbyte(height)); err !=nil {
		return err
	}
	return nil
}

func (b *badgerDB) updateScoreTX(pubKey []byte, tx *badger.Txn) error {
	// 积分键
	scoreKey := getScoreKey(pubKey)
	item, err := tx.Get(scoreKey)
	if err != nil && err != badger.ErrKeyNotFound {
		return err
	}
	origin := uint64(0)
	if err != badger.ErrKeyNotFound {
		item.Value(func(val []byte) error {
			origin = byteh(val)
			return nil
		})
	}
	origin ++
	return tx.Set(scoreKey, hbyte(origin))
}

func (b *badgerDB) updateLatestHeightTX() {}

func (b *badgerDB) view(fn func(txn *badger.Txn) error) error {
	return b.wrapError(b.View(fn))
}

func (b *badgerDB) update(fn func(txn *badger.Txn) error) error {
	return b.wrapError(b.update(fn))
}

// 包装错误直接从badger
func (b *badgerDB) wrapError(err error) error {
	if err == nil {
		return nil
	}
	if err == badger.ErrKeyNotFound {
		return ErrNotFound
	}
	logger.Warn("badger got unexpect err: %v\n", err)
	return ErrInternal
}


func (b *badgerDB) start() {
	go b.gcloop()
	b.lm.StartWorking()
}

func (b *badgerDB) stop() {
	b.lm.Stop()
}

func (b *badgerDB) gcloop() {
	b.lm.Add()
	defer b.lm.Done()

	ticker := time.NewTicker(time.Minute *10)

	for {
		select {

		case <- b.lm.D:
			return
			case <- ticker.C:
				b.RunValueLogGc(0.5)
		}
	}
}