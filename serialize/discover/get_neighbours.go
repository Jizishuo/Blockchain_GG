package discover

type GetNeighbours struct {
	*Head
	PubKey []byte
}

func NewGetNeighbours(pubKey []byte) *GetNeighbours {
	return &GetNeighbours{
		Head: NewHeadV1(MsgGetNeighbours),
		PubKey: pubKey,
	}
}
