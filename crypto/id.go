package crypto

import (
	"encoding/base32"
	"github.com/btcsuite/btcd/btcec"
)

var (
	base32Codec = base32.StdEncoding.WithPadding(base32.NoPadding)
)
func PubKeyToID(pubKey *btcec.PublicKey) string {
	pubKeyB := pubKey.SerializeCompressed()
	return base32Codec.EncodeToString(pubKeyB)
}
// PrivKeyToID returns a peer id from the private(私人) key
func PrivKeyToID(privKey *btcec.PrivateKey) string {
	pubKey := privKey.PubKey()
	return PubKeyToID(pubKey)
}