package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ConflictResolver 文件名冲突解决器接口
type ConflictResolver interface {
	// Resolve 检查路径是否冲突，返回实际路径、是否重命名、重命名计数
	Resolve(path string) (actualPath string, wasRenamed bool, renameCount int)
	Reset() // 重置内部状态
}

// conflictResolver 基于计数器的冲突解决器
type conflictResolver struct {
	usedPaths map[string]int // 路径 -> 使用次数
	mu        sync.Mutex     // 并发保护
}

// NewConflictResolver 创建冲突解决器
func NewConflictResolver() ConflictResolver {
	return &conflictResolver{
		usedPaths: make(map[string]int),
	}
}

func (r *conflictResolver) Resolve(path string) (string, bool, int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查路径是否已被使用
	count := r.usedPaths[path]
	r.usedPaths[path] = count + 1

	if count == 0 && !r.fileExists(path) {
		// 路径未被使用且文件不存在，直接使用
		return path, false, 0
	}

	// 需要重命名
	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, ext)

	// 生成带后缀的新文件名
	renameCount := count
	for {
		newName := fmt.Sprintf("%s_conflict%d%s", base, renameCount, ext)
		newPath := filepath.Join(dir, newName)

		if r.usedPaths[newPath] == 0 && !r.fileExists(newPath) {
			r.usedPaths[newPath] = 1
			return newPath, true, renameCount
		}

		renameCount++
		// 安全检查，防止无限循环
		if renameCount > 10000 {
			// 使用时间戳作为后缀
			newName = fmt.Sprintf("%s_%d%s", base, time.Now().UnixNano(), ext)
			return filepath.Join(dir, newName), true, renameCount
		}
	}
}

func (r *conflictResolver) fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (r *conflictResolver) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.usedPaths = make(map[string]int)
}
