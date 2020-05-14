package cp

import (
	"math/big"
)

// powCache 用于挖掘，以减少调用 Marshal 的时间（）;
// 结果只能读，不应修改
type powCache struct {
	marshalCache []byte
	powCache *big.Int
	cache bool
}

func newPowCache() *powCache {
	return &powCache{
		cache: false,
	}
}