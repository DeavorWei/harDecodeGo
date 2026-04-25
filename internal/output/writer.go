package output

import (
	"fmt"
	"os"
	"path/filepath"
)

// Writer 文件写入器接口
type Writer interface {
	Write(data []byte, path string) error
	Exists(path string) bool
	MkdirAll(path string) error
}

type writer struct{}

// NewWriter 创建写入器
func NewWriter() Writer {
	return &writer{}
}

func (w *writer) Write(data []byte, path string) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := w.MkdirAll(dir); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (w *writer) Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (w *writer) MkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}
