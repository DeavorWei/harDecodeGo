package extractor

import (
	"path/filepath"
	"strings"
)

// MimeTypeMapper MIME类型映射器接口
type MimeTypeMapper interface {
	GetExtension(mimeType string) string
	Match(mimeType, pattern string) bool
}

type mimeTypeMapper struct {
	mimeToExt map[string]string
}

// NewMimeTypeMapper 创建MIME映射器
func NewMimeTypeMapper() MimeTypeMapper {
	return &mimeTypeMapper{
		mimeToExt: map[string]string{
			// 文本类型
			"text/html":       ".html",
			"text/css":        ".css",
			"text/javascript": ".js",
			"text/plain":      ".txt",
			"text/xml":        ".xml",
			"text/csv":        ".csv",
			"text/markdown":   ".md",

			// 应用类型
			"application/javascript":   ".js",
			"application/json":         ".json",
			"application/xml":          ".xml",
			"application/pdf":          ".pdf",
			"application/zip":          ".zip",
			"application/gzip":         ".gz",
			"application/wasm":         ".wasm",
			"application/octet-stream": ".bin",

			// 图片类型
			"image/jpeg":               ".jpg",
			"image/jpg":                ".jpg",
			"image/png":                ".png",
			"image/gif":                ".gif",
			"image/webp":               ".webp",
			"image/svg+xml":            ".svg",
			"image/x-icon":             ".ico",
			"image/vnd.microsoft.icon": ".ico",
			"image/bmp":                ".bmp",
			"image/tiff":               ".tiff",

			// 字体类型
			"font/woff":              ".woff",
			"font/woff2":             ".woff2",
			"font/ttf":               ".ttf",
			"font/otf":               ".otf",
			"application/font-woff":  ".woff",
			"application/font-woff2": ".woff2",
			"application/x-font-ttf": ".ttf",
			"application/x-font-otf": ".otf",

			// 视频/音频类型
			"video/mp4":  ".mp4",
			"video/webm": ".webm",
			"video/ogg":  ".ogv",
			"audio/mpeg": ".mp3",
			"audio/wav":  ".wav",
			"audio/ogg":  ".oga",
			"audio/mp4":  ".m4a",

			// 文档类型
			"application/msword": ".doc",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
			"application/vnd.ms-excel": ".xls",
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         ".xlsx",
			"application/vnd.ms-powerpoint":                                             ".ppt",
			"application/vnd.openxmlformats-officedocument.presentationml.presentation": ".pptx",
		},
	}
}

func (m *mimeTypeMapper) GetExtension(mimeType string) string {
	// 清理MIME类型（移除charset等参数）
	cleanMime := strings.Split(mimeType, ";")[0]
	cleanMime = strings.TrimSpace(cleanMime)
	cleanMime = strings.ToLower(cleanMime)

	if ext, ok := m.mimeToExt[cleanMime]; ok {
		return ext
	}
	return ""
}

func (m *mimeTypeMapper) Match(mimeType, pattern string) bool {
	// 清理MIME类型
	cleanMime := strings.Split(mimeType, ";")[0]
	cleanMime = strings.TrimSpace(cleanMime)
	cleanMime = strings.ToLower(cleanMime)

	pattern = strings.ToLower(strings.TrimSpace(pattern))

	// 支持通配符匹配，如 "image/*"
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(cleanMime, prefix+"/")
	}
	return cleanMime == pattern
}

// GetExtensionFromURL 从URL路径推断扩展名
func GetExtensionFromURL(urlStr string) string {
	ext := filepath.Ext(urlStr)
	if ext != "" {
		return ext
	}
	return ""
}
