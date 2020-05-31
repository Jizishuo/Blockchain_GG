package crypto

import (
	"Blockchain_GG/utils"
	"errors"
	"github.com/btcsuite/btcd/btcec"
	"io/ioutil"
	"strings"
)

/*
	Pkey 是存储在磁盘上的普通私钥。
*/
const (
	PlainKeyType = 1
	PlainKey     = ".pKey"
)

// 从文件还原私钥还原密钥
func RestorePKey(path string) (*btcec.PrivateKey, error) {
	keyFile := path + "/" + PlainKey
	hexPrivKey, err := readKeyFile(keyFile)
	if err != nil {
		return nil, err
	}
	bytePrivKey, err := utils.FromHex(string(hexPrivKey))
	if err != nil {
		return nil, err
	}
	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), bytePrivKey)
	if privKey == nil {
		return nil, errors.New("parse bytes to privkey is faild")
	}
	return privKey, nil
}

func readKeyFile(file string) ([]byte, error) {
	if err := utils.AccessCheck(file); err != nil {
		return nil, err
	}
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	// 前后端所有空白（unicode.IsSpace指定）都去掉的字符串。
	trimContent := strings.TrimSpace(string(content))
	return []byte(trimContent), nil
}
