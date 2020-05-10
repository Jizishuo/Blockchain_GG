package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
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

// ParseIPPort parse IP:Port format sring
func ParseUPPort(ipPort string) (net.IP, int) {
	s := strings.Split(ipPort, ":")
	if len(s) != 2 {
		return nil, 0
	}
	// ParseIP将s解析为IP地址，并返回该地址。如果s不是合法的IP地址文本表示，ParseIP会返回nil
	ip := net.ParseIP(s[0])
	if ip == nil || ip.To4()==nil {
		return nil, 0
	}
	// 是ParseInt(s, 10, 0)的简写
	port, err := strconv.Atoi(s[1])
	if err != nil || port <= 0 || port > 65535 {
		return nil, 0
	}
	return ip, port
}