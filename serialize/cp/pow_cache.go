package cp

import (
	"Blockchain_GG/utils"
	"encoding/binary"
	"math/big"
)

// powCache 用于挖掘，以减少调用 Marshal 的时间（）;
// 结果只能读，不应修改
type powCache struct {
	marshalCache []byte
	powCache     *big.Int
	cache        bool
}

func newPowCache() *powCache {
	return &powCache{
		cache: false,
	}
}

func (p *powCache) cacheBefore() bool {
	return p.cache
}

func (p *powCache) setCache(marshal []byte, pow *big.Int) {
	p.marshalCache = marshal
	p.powCache = pow
	p.cache = true
}

func (p *powCache) update(nonce uint32, index int) *big.Int {
	binary.BigEndian.PutUint32(p.marshalCache[index:], nonce)
	return p.powCache.SetBytes(utils.Hash(p.marshalCache))
}
