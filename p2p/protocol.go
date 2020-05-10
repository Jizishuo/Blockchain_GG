package p2p


// Protocol is the interface that the p2p network protocols must implement
type Protocol interface {
	ID() uint8
	Name() string
}

type ProtocolRunner interface {
	Send(dp *P)
}