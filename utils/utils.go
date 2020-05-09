package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
)

// 检测文件是否存在
func AccessCheck(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("not found %s or permision denied", err)
	}
	return nil
}

// hex编码的字符串s代表的数据
func FromHex(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

var logger = &Logger{
	Logger:log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
}

// Uint8Len returns bytes length in uint8 type
func Uint8Len(data []byte) uint8 {
	return uint8(len(data))
}
// ToHex returns the upper case hexadecimal encoding string
func ToHex(data []byte) string {
	return strings.ToUpper(hex.EncodeToString(data))
}

func Hash(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}