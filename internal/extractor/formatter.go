package extractor

import (
	"fmt"
	"strings"

	"har-decode/internal/har"
)

// HTTPFormatter HTTP格式化器，用于将HAR Entry格式化为可读的HTTP请求/响应文本
type HTTPFormatter struct{}

// NewHTTPFormatter 创建HTTP格式化器
func NewHTTPFormatter() *HTTPFormatter {
	return &HTTPFormatter{}
}

// FormatFullHTTP 格式化完整HTTP请求/响应信息
// 包含：请求行、请求头、查询参数、响应状态、响应头、响应内容
func (f *HTTPFormatter) FormatFullHTTP(entry *har.Entry, body string) string {
	var sb strings.Builder

	// ========== 请求部分 ==========
	sb.WriteString("========== HTTP REQUEST ==========\n\n")

	// 请求行：METHOD URL HTTPVersion
	sb.WriteString(fmt.Sprintf("%s %s %s\n",
		entry.Request.Method,
		entry.Request.URL,
		entry.Request.HTTPVersion))

	// 请求头
	sb.WriteString("\n--- Request Headers ---\n")
	for _, h := range entry.Request.Headers {
		sb.WriteString(fmt.Sprintf("%s: %s\n", h.Name, h.Value))
	}

	// 查询参数
	if len(entry.Request.QueryString) > 0 {
		sb.WriteString("\n--- Query Parameters ---\n")
		for _, q := range entry.Request.QueryString {
			sb.WriteString(fmt.Sprintf("%s: %s\n", q.Name, q.Value))
		}
	}

	// 请求Cookies（如果有）
	if len(entry.Request.Cookies) > 0 {
		sb.WriteString("\n--- Request Cookies ---\n")
		for _, c := range entry.Request.Cookies {
			sb.WriteString(fmt.Sprintf("%s: %s\n", c.Name, c.Value))
		}
	}

	// 请求体大小（如果有）
	if entry.Request.BodySize > 0 {
		sb.WriteString(fmt.Sprintf("\n--- Request Body (%d bytes) ---\n", entry.Request.BodySize))
	}

	// ========== 响应部分 ==========
	sb.WriteString("\n========== HTTP RESPONSE ==========\n\n")

	// 状态行：HTTPVersion StatusCode StatusText
	sb.WriteString(fmt.Sprintf("%s %d %s\n",
		entry.Response.HTTPVersion,
		entry.Response.Status,
		entry.Response.StatusText))

	// 响应头
	sb.WriteString("\n--- Response Headers ---\n")
	for _, h := range entry.Response.Headers {
		sb.WriteString(fmt.Sprintf("%s: %s\n", h.Name, h.Value))
	}

	// 响应Cookies（如果有）
	if len(entry.Response.Cookies) > 0 {
		sb.WriteString("\n--- Response Cookies ---\n")
		for _, c := range entry.Response.Cookies {
			sb.WriteString(fmt.Sprintf("%s: %s\n", c.Name, c.Value))
		}
	}

	// 响应内容
	sb.WriteString("\n--- Response Body ---\n")
	if body != "" {
		sb.WriteString(body)
		// 确保末尾有换行
		if !strings.HasSuffix(body, "\n") {
			sb.WriteString("\n")
		}
	} else {
		sb.WriteString("[Empty]\n")
	}

	return sb.String()
}
