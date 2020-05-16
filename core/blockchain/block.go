package blockchain

import (
	"Blockchain_GG/serialize/cp"
	"Blockchain_GG/utils"
	"fmt"
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
		hash: b.GetSerializedHash(),
		height:height,
		stored: stored,
	}
}
func (b *block) target() uint32 {
	return b.Target
}

func (b *block) time() int64 {
	return b.Time
}

func (b *block) isStored() bool {
	return b.stored
}

func (b *block) setBackward(back *block) {
	b.backward = back
}

func (b *block) removeBackward() {
	b.backward = nil
}

func (b *block) addFordward(forward *block) {
	key := utils.ToHex(forward.hash)
	b.fordward.Store(key, forward)
}

func (b *block) removeForward(forward *block) {
	key := utils.ToHex(forward.hash)
	b.fordward.Delete(key)
}

func (b *block) isBackwardOf(cb *cp.Block) bool {
	key := utils.ToHex(cb.GetSerializedHash()) // TODO
	// key := utils.ToHex(cb.GetSerializeHash())
	_, ok := b.fordward.Load(key)
	return ok
}

func (b *block) forwardNum() int {
	result := 0
	b.fordward.Range(func(k, v interface{}) bool {
		result++
		return true
	})
	return result
}

// remove the block from its backward block, and returns the backward block;
// the block should never use after removing
func (b *block) remove() (*block, error) {
	if b.forwardNum() != 0 {
		return nil, fmt.Errorf("fordward reference is not zero, can't be removed")
	}

	backward := b.backward
	backward.removeForward(b)
	b.removeBackward()
	return backward, nil
}
