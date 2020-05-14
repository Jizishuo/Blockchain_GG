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
	"time"
)

const (
	HashLength = sha256.Size
	timeFormat = "2020/02/02 02:02:02"
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
// Uint16Len returns bytes length in uint16 type
func Uint16Len(data []byte) uint16 {
	return uint16(len(data))
}

// Uint32Len returns bytes length in uint32 type
func Uint32Len(data []byte) uint32 {
	return uint32(len(data))
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
//TimeToString 返回时间的文本表示形式;
//它只接受int64或时间。时间类型
func TimeToString(t interface{}) string {
	if int64T, ok := t.(int64); ok {
		return time.Unix(int64T, 0).Format(timeFormat)
	}
	if timeT, ok:= t.(time.Time);ok {
		return timeT.Format(timeFormat)
	}
	logger.Fatal("invalid call to timetostring (%v)\n", t)
	return ""
}