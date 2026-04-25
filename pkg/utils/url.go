package utils

import (
	"net/url"
	"strings"
)

// ExtractHost 从URL提取主机名
func ExtractHost(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return u.Host, nil
}

// IsValidURL 检查URL是否有效
func IsValidURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}

// ExtractPath 从URL提取路径部分
func ExtractPath(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return u.Path, nil
}

// DecodeURL URL解码
func DecodeURL(encoded string) string {
	decoded, err := url.PathUnescape(encoded)
	if err != nil {
		return encoded
	}
	return decoded
}

// GetQueryParams 获取URL查询参数
func GetQueryParams(rawURL string) (map[string]string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	params := make(map[string]string)
	for key, values := range u.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}
	return params, nil
}

// RemoveQueryString 移除URL查询字符串
func RemoveQueryString(rawURL string) string {
	idx := strings.Index(rawURL, "?")
	if idx != -1 {
		return rawURL[:idx]
	}
	return rawURL
}
