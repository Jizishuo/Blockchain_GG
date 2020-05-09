package peer

import "time"

var (
	initTimepoint = time.Unix(0,0)
)

type pstate struct {
	*Peer
	// 种子不应该删除，一旦他们被添加到peer
	isSeed bool
	hasPingBefore bool
	lastActiveTime time.Time
	lastGetNeighbourTime time.Time
}

func newPState(p *Peer, isSeed bool) *pstate {
	return &pstate{
		Peer:p,
		isSeed: isSeed,
		hasPingBefore: false,
		lastActiveTime: initTimepoint,
		lastGetNeighbourTime: initTimepoint,
	}
}