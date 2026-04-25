package output

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"har-decode/internal/logger"
	"har-decode/pkg/utils"
)

const (
	MaxPathLength        = 250           // Windows MAX_PATH限制为260，留余量
	ConflictSuffixFormat = "_conflict%d" // 冲突文件名后缀格式
)

// PathResult 路径构建结果
type PathResult struct {
	OriginalPath string // 原始期望路径
	ActualPath   string // 实际使用路径
	WasRenamed   bool   // 是否发生重命名
	RenameCount  int    // 重命名次数
}

// MimeTypeMapper MIME类型映射器接口（从extractor包引用）
type MimeTypeMapper interface {
	GetExtension(mimeType string) string
	Match(mimeType, pattern string) bool
}

// PathBuilder 路径构建器接口
type PathBuilder interface {
	Build(resourceURL, mimeType, outputDir string) (*PathResult, error)
}

type pathBuilder struct {
	mimeMapper       MimeTypeMapper
	conflictResolver ConflictResolver
	logger           logger.Logger
}

// NewPathBuilder 创建路径构建器
func NewPathBuilder(mimeMapper MimeTypeMapper, resolver ConflictResolver, log logger.Logger) PathBuilder {
	return &pathBuilder{
		mimeMapper:       mimeMapper,
		conflictResolver: resolver,
		logger:           log,
	}
}

func (b *pathBuilder) Build(resourceURL, mimeType, outputDir string) (*PathResult, error) {
	result := &PathResult{}

	// 解析URL
	parsedURL, err := url.Parse(resourceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// 构建基础路径
	var relativePath string
	if parsedURL.Path == "" || parsedURL.Path == "/" {
		// 根路径使用index.html
		relativePath = "index.html"
	} else {
		relativePath = parsedURL.Path

		// 移除开头的斜杠
		relativePath = strings.TrimPrefix(relativePath, "/")

		// URL解码
		decodedPath, err := url.PathUnescape(relativePath)
		if err == nil {
			relativePath = decodedPath
		}
	}

	// 清理文件名中的非法字符
	relativePath = utils.SanitizeFilePath(relativePath)

	// 如果没有扩展名，根据MIME类型添加
	if filepath.Ext(relativePath) == "" && mimeType != "" {
		ext := b.mimeMapper.GetExtension(mimeType)
		if ext != "" {
			relativePath += ext
		}
	}

	// 组合完整路径
	fullPath := filepath.Join(outputDir, relativePath)
	result.OriginalPath = fullPath

	// 检查路径长度限制
	if len(fullPath) > MaxPathLength {
		// 尝试缩短路径：使用哈希替代文件名
		shortenedPath := b.shortenPath(fullPath, MaxPathLength)
		b.logger.Warn("Path too long, shortened",
			logger.F("original", fullPath),
			logger.F("shortened", shortenedPath))
		fullPath = shortenedPath
	}

	// 处理文件名冲突
	actualPath, renamed, count := b.conflictResolver.Resolve(fullPath)
	result.ActualPath = actualPath
	result.WasRenamed = renamed
	result.RenameCount = count

	if renamed {
		b.logger.Debug("File name conflict resolved",
			logger.F("original", fullPath),
			logger.F("actual", actualPath))
	}

	return result, nil
}

func (b *pathBuilder) shortenPath(path string, maxLen int) string {
	// 使用MD5哈希缩短文件名
	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, ext)

	hash := utils.HashStringShort(base)
	newName := hash + ext

	newPath := filepath.Join(dir, newName)
	if len(newPath) > maxLen && len(dir) > 10 {
		// 如果仍然太长，缩短目录名
		dir = dir[:maxLen-len(newName)-10] + "..."
		newPath = filepath.Join(dir, newName)
	}

	return newPath
}
