package crypto

/*
	密钥是存储在磁盘上的密封私钥。
	它比普通密钥存储更安全。用户可以从密钥导出密钥，反之亦然。
	用于加密 ecc 私钥的 aes 密钥由 scrypt 派生。
*/
const (
	SealKeyType = 2
	SealKey     = ".skey"

	version    = 1
	kdfName    = "scrypt"
	dkfLen     = 32
	scryptN    = 262144
	scryptP    = 1
	scryptR    = 8
	saltLen    = 32
	crtpyoName = "aes-256-gcm"
)
