package utils

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
)

// FileExists 检查文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsDir 检查路径是否为目录
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// EnsureDir 确保目录存在
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// HashString 计算字符串哈希（用于缩短长文件名）
func HashString(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

// HashStringShort 计算字符串短哈希（前8位）
func HashStringShort(s string) string {
	return HashString(s)[:8]
}

// GetFileSize 获取文件大小
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// TruncatePath 截断路径到指定长度
func TruncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}

	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, ext)

	// 使用哈希缩短文件名
	hash := HashStringShort(base)
	newName := hash + ext

	newPath := filepath.Join(dir, newName)
	if len(newPath) > maxLen && len(dir) > 10 {
		// 如果仍然太长，缩短目录名
		dir = dir[:maxLen-len(newName)-10] + "..."
		newPath = filepath.Join(dir, newName)
	}

	return newPath
}
