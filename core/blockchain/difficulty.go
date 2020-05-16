package blockchain

import (
	"Blockchain_GG/utils"
	"math/big"
	"time"
	"fmt"

)


const (
	//	(difficulty target)难度目标是 uint32 类型（4 字节）
	//	第一个（32-'diffPrefixBitsLen'）字节表示难度的位长度;
	//	其余部分是难度的前缀
	diffPrefixBitsLen = 24

	// ReferenceBlocks sets how many previous blocks to adjust difficluty
	//参考块设置要调整差异的先前块数
	ReferenceBlocks = 20

	LastDurationWeight     = float64(1) / float64(alpha)
	PreviousDurationWeight = float64(1) - float64(LastDurationWeight)
)

var (
	BlockTargetLimit        uint32
	EvidenceTargetLimit     uint32
	BlockDifficultyLimit    *big.Int
	EvidenceDifficultyLimit *big.Int

	// BlockInterval expects every new block generation interval
	// 块间隔期望每个新的块生成间隔
	BlockInterval           time.Duration
	floatBlockInterval      *big.Float
	expectReferenceInterval time.Duration
)

func initMiningParams(conf *Config) {
	BlockTargetLimit = conf.BlockTargetLimit
	EvidenceTargetLimit = conf.EvidenceTargetLimit
	BlockDifficultyLimit = TargetToDiff(BlockTargetLimit)
	EvidenceDifficultyLimit = TargetToDiff(EvidenceTargetLimit)

	BlockInterval = time.Duration(conf.BlockInterval) * time.Second
	floatBlockInterval = big.NewFloat(float64(BlockInterval))
	expectReferenceInterval = BlockInterval * ReferenceBlocks

	logger.Info("initialize mining params: BlockDifficultyLimit %s, EvidenceDifficultyLimit %s, interval %v",
		utils.ReadableBigInt(BlockDifficultyLimit), utils.ReadableBigInt(EvidenceDifficultyLimit), BlockInterval)
}

// CalculateTarget calculates latest targest
// lastDuration is the interval between now and the latest block
// preDuration is the interval between the latest block and the 20 blocks before it
//计算目标计算最新的焦油
//最后持续时间是从现在到最新块之间的间隔
//前持续时间是最新块和之前 20 个块之间的间隔
func CalculateTarget(lastTarget uint32, lastDuration time.Duration, preDuration time.Duration) uint32 {
	lastDiff := TargetToDiff(lastTarget)
	floatLastDiff := new(big.Float).SetInt(lastDiff)

	if lastDuration < 1*time.Second {
		lastDuration = 1 * time.Second
	}

	averageInterval := big.NewFloat(float64(lastDuration)*LastDurationWeight +
		float64(preDuration/ReferenceBlocks)*PreviousDurationWeight)
	scale := new(big.Float).Quo(averageInterval, floatBlockInterval)
	floatCurrDiff := new(big.Float).Mul(floatLastDiff, scale)

	// check the lower bound 检查下限
	const lowerBoundScale = 0.5
	lowerBoundDiff := new(big.Float).Mul(floatLastDiff, big.NewFloat(lowerBoundScale))
	if floatCurrDiff.Cmp(lowerBoundDiff) < 0 {
		logger.Debug("trigger lower bound, scale change to 0.5")
		scale.SetFloat64(lowerBoundScale)
		floatCurrDiff = lowerBoundDiff
	}

	intCurrDiff := new(big.Int)
	floatCurrDiff.Int(intCurrDiff)

	// check the limit 检查限制
	if intCurrDiff.Cmp(BlockDifficultyLimit) > 0 {
		return BlockTargetLimit
	}

	logger.Debug("past %d blocks use %v, expect %v, %v away last block, scale %.2f, get diff %s\n",
		ReferenceBlocks, preDuration, expectReferenceInterval, lastDuration,
		scale, utils.ReadableBigInt(intCurrDiff))

	currTarget := DiffToTarget(intCurrDiff)
	return currTarget
}

// TargetToDiff transforms 32 bits target to 256 bits difficulty
//TargetToDiff 将 32 位目标转换为 256 位困难
func TargetToDiff(target uint32) *big.Int {

	// the number of difficulty bits;
	// the difficulty may begin with 0, like 00001000, witch is equal to 1000 but its length is 8 instead of 4;
	// so the length is not the final length of bits in big.Int (it may not equal to big.Int.BitLen())
	//	难度位数;
	//	难度可能从 0 开始，如 00001000，巫婆等于 1000，但其长度为 8 而不是 4;
	//	所以长度不是大位的最终长度。int（它可能不等于大。In.BitLen（））
	length := target >> diffPrefixBitsLen

	// difficulty prefix 难度前缀
	prefix := target & 0x00FFFFFF

	diff := big.NewInt(int64(prefix))
	if length > diffPrefixBitsLen {
		lsh := length - diffPrefixBitsLen
		diff = diff.Lsh(diff, uint(lsh))
	}

	return diff
}

// DiffToTarget transforms 256 bits difficulty to 32 bits target
// DiffToTarget 将 256 位难度转换为 32 位目标
func DiffToTarget(diff *big.Int) uint32 {
	var target uint32

	binaryBits := fmt.Sprintf("%b", diff)
	if binaryBits == "0" {
		return 0
	}

	length := uint32(len(binaryBits))
	target = length << diffPrefixBitsLen

	if length > diffPrefixBitsLen {
		diff = diff.Rsh(diff, uint(length-diffPrefixBitsLen))
	}

	target = target | uint32(diff.Int64())
	return target
}
