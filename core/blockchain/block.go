package blockchain

import (
	"Blockchain_GG/serialize/cp"
	"sync"
)

type block struct {
	*cp.Block
	hash []byte
	height uint64
	stored bool  // 储存
	// 向后块是此块的父块，只有一个
	backward *block

	// 	前进块是此块的子块。
	//	如果有多个孩子，则意味着分叉正在发生。
	//	[字符串，[块]，十六进制（哈希）作为键  <string, *block>, hex(hash) as key
	fordward sync.Map
}

func newBlock(b *cp.Block, height uint64, stored bool) *block {
	return &block{
		Block:b,
		hash: b.GetSerializeHash(),
		height:height,
		stored: stored,
	}
}