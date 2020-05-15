package blockchain

import "Blockchain_GG/utils"

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
	PassiveChangeNotify chan bool

	oldestBlock *block
}