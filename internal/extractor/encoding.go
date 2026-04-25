package extractor

import (
	"encoding/base64"
	"fmt"

	"har-decode/internal/har"
)

// EncodingType 编码类型
type EncodingType string

const (
	EncodingBase64 EncodingType = "base64"
	EncodingRaw    EncodingType = ""
)

// ContentDecoder 内容解码器接口
type ContentDecoder interface {
	Decode(content *har.Content) ([]byte, error)
	SupportsEncoding(encoding string) bool
}

type contentDecoder struct {
	supportedEncodings map[string]bool
}

// NewContentDecoder 创建解码器
func NewContentDecoder() ContentDecoder {
	return &contentDecoder{
		supportedEncodings: map[string]bool{
			"base64": true,
			"":       true, // 无编码（纯文本）
		},
	}
}

func (d *contentDecoder) SupportsEncoding(encoding string) bool {
	return d.supportedEncodings[encoding]
}

func (d *contentDecoder) Decode(content *har.Content) ([]byte, error) {
	// 检查空内容
	if content.Text == "" {
		return nil, har.NewError(har.ErrEmptyContent, "content text is empty", nil)
	}

	// 根据编码类型解码
	switch content.Encoding {
	case "base64":
		decoded, err := base64.StdEncoding.DecodeString(content.Text)
		if err != nil {
			return nil, har.NewError(har.ErrDecodeFailed, "failed to decode base64 content", err)
		}
		return decoded, nil

	case "":
		// 纯文本，直接返回
		return []byte(content.Text), nil

	default:
		return nil, har.NewError(har.ErrDecodeFailed, fmt.Sprintf("unsupported encoding: %s", content.Encoding), nil)
	}
}
