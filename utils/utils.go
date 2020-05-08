package utils

import (
	"fmt"
	"os"
)

// 检测文件是否存在
func AccessCheck(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("not found %s or permision denied", err)
	}
	return nil
}

