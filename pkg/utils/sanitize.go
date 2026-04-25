package utils

import (
	"path/filepath"
	"strings"
)

// Windows非法字符
var invalidChars = []string{"<", ">", ":", "\"", "/", "\\", "|", "?", "*"}

// SanitizeFileName 清理文件名中的非法字符
func SanitizeFileName(name string) string {
	result := name
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}
	// 移除控制字符
	result = strings.Map(func(r rune) rune {
		if r < 32 {
			return '_'
		}
		return r
	}, result)

	// 移除首尾空格和点
	result = strings.TrimSpace(result)
	result = strings.Trim(result, ".")

	// 处理空文件名
	if result == "" {
		result = "unnamed"
	}

	return result
}

// SanitizeFilePath 清理文件路径
func SanitizeFilePath(path string) string {
	parts := strings.Split(path, string(filepath.Separator))
	for i, part := range parts {
		parts[i] = SanitizeFileName(part)
	}
	return strings.Join(parts, string(filepath.Separator))
}
