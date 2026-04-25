package har

import (
	"encoding/json"
	"io"
	"os"

	"har-decode/internal/logger"
)

// Parser HAR解析器接口
type Parser interface {
	// Parse 从文件路径解析HAR（适合小文件）
	Parse(filePath string) (*HAR, error)
	// ParseFromBytes 从字节数组解析（适合已加载到内存的数据）
	ParseFromBytes(data []byte) (*HAR, error)
	// ParseStream 流式解析（适合大文件，内存友好）
	ParseStream(filePath string, entryHandler func(*Entry) error) error
}

type parser struct {
	logger logger.Logger
}

// NewParser 创建解析器实例
func NewParser(log logger.Logger) Parser {
	return &parser{logger: log}
}

func (p *parser) Parse(filePath string) (*HAR, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, NewError(ErrInvalidFile, "failed to read HAR file", err)
	}
	return p.ParseFromBytes(data)
}

func (p *parser) ParseFromBytes(data []byte) (*HAR, error) {
	var har HAR
	if err := json.Unmarshal(data, &har); err != nil {
		return nil, NewError(ErrParseFailed, "failed to parse HAR JSON", err)
	}
	p.logger.Debug("HAR parsed successfully",
		logger.F("entries", len(har.Log.Entries)))
	return &har, nil
}

func (p *parser) ParseStream(filePath string, entryHandler func(*Entry) error) error {
	file, err := os.Open(filePath)
	if err != nil {
		return NewError(ErrInvalidFile, "failed to open HAR file", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	// 定位到entries数组
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return NewError(ErrParseFailed, "failed to read token", err)
		}
		if str, ok := token.(string); ok && str == "entries" {
			// 读取数组开始标记 [
			_, err = decoder.Token()
			if err != nil {
				return NewError(ErrParseFailed, "failed to read array start", err)
			}
			break
		}
	}

	// 流式处理每个entry
	var entryCount int
	for decoder.More() {
		var entry Entry
		if err := decoder.Decode(&entry); err != nil {
			if err == io.EOF {
				break
			}
			return NewError(ErrParseFailed, "failed to decode entry", err)
		}
		entryCount++
		if err := entryHandler(&entry); err != nil {
			return err
		}
	}

	p.logger.Info("Stream parsing completed",
		logger.F("entries_processed", entryCount))
	return nil
}
