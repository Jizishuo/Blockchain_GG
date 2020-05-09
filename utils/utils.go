package utils

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
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