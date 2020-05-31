package p2p

// 编解码器用于加密/解密消息
type codec interface {
	encrypt(plainText []byte) ([]byte, error)
	decrypt(cipherText []byte) ([]byte, error)
}
