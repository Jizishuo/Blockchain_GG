package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	HashLength = sha256.Size
	timeFormat = "2020/02/02 02:02:02"
)

var logger = &Logger{
	Logger:log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
}


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
var bufPool = &sync.Pool{
	// Pool是一个可以分别存取的临时对象的集合
	New: func() interface {} {
		return new(bytes.Buffer)
	},
}
// GetBuf gets a *bytes.Buffer from pool
func GetBuf() *bytes.Buffer {
	result := bufPool.Get().(*bytes.Buffer)
	result.Reset() //Reset重设缓冲，因此会丢弃全部内容，等价于b.Truncate(0)
	return result
}
// ReturnBuf returns a *bytes.Buffer to Pool once you don't need it
func ReturnBuf(buf *bytes.Buffer) {
	// Put方法将x放入池中
	bufPool.Put(buf)
}

// ReadableBigInt returns more readable format for big.Int
// 可读 BigInt 返回大格式的可读格式。Int
func ReadableBigInt(num *big.Int) string {
	hexStr := fmt.Sprintf("%X", num)
	length := len(hexStr)

	var result string
	format := "0x%s..(%d)"
	cut := 6
	if length > cut {
		result = fmt.Sprintf(format, hexStr[0:cut], length)
	} else {
		result = fmt.Sprintf(format, hexStr, length)
	}
	return result
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