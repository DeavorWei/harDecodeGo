package extractor

import (
	"sort"
	"strconv"
	"sync"

	"har-decode/internal/har"
	"har-decode/internal/logger"
	"har-decode/internal/output"
)

// ExtractionStrategy 提取策略
type ExtractionStrategy int

const (
	StrategyContinueOnError  ExtractionStrategy = iota // 遇到错误继续处理后续entries
	StrategyStopOnFirstError                           // 遇到第一个错误立即停止
	StrategySkipEmptyContent                           // 跳过空内容条目
)

// ExtractConfig 提取配置
type ExtractConfig struct {
	OutputDir         string             // 输出目录
	Strategy          ExtractionStrategy // 提取策略
	Workers           int                // 并发worker数量，0表示使用CPU核心数
	FilterMimeTypes   []string           // MIME类型白名单（空表示全部）
	FilterStatusCodes []int              // HTTP状态码白名单（空表示全部）
	Verbose           bool               // 详细输出
}

// ExtractResult 单个entry的提取结果
type ExtractResult struct {
	URL        string
	OutputPath string
	MimeType   string
	Size       int
	StatusCode int
	Success    bool
	Skipped    bool // 是否被过滤器跳过
	Error      error
}

// ExtractReport 提取报告
type ExtractReport struct {
	TotalEntries  int
	SuccessCount  int
	FailedCount   int
	SkippedCount  int
	FilteredCount int // 被过滤器排除的数量
	Results       []ExtractResult
	Errors        []ExtractError
}

// ExtractError 提取错误
type ExtractError struct {
	URL   string
	Phase string // parse/decode/write
	Error error
}

// IndexedTask 带序号的提取任务（用于并发处理）
type IndexedTask struct {
	Entry  *har.Entry
	Index  int // 序号（从1开始）
	Digits int // 序号位数
}

// Extractor 提取器接口
type Extractor interface {
	Extract(harData *har.HAR, config *ExtractConfig) (*ExtractReport, error)
	ExtractParallel(harData *har.HAR, config *ExtractConfig) (*ExtractReport, error)
}

type extractor struct {
	writer      output.Writer
	decoder     ContentDecoder
	pathBuilder output.PathBuilder
	mimeMapper  MimeTypeMapper
	formatter   *HTTPFormatter
	logger      logger.Logger
}

// NewExtractor 创建提取器
func NewExtractor(
	writer output.Writer,
	decoder ContentDecoder,
	pathBuilder output.PathBuilder,
	mimeMapper MimeTypeMapper,
	log logger.Logger,
) Extractor {
	return &extractor{
		writer:      writer,
		decoder:     decoder,
		pathBuilder: pathBuilder,
		mimeMapper:  mimeMapper,
		formatter:   NewHTTPFormatter(),
		logger:      log,
	}
}

// calculateDigits 计算序号需要的位数
func calculateDigits(total int) int {
	if total <= 0 {
		return 1
	}
	return len(strconv.Itoa(total))
}

// sortEntriesByTime 按请求时间排序entries
func sortEntriesByTime(entries []har.Entry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].StartedDateTime < entries[j].StartedDateTime
	})
}

func (e *extractor) Extract(harData *har.HAR, config *ExtractConfig) (*ExtractReport, error) {
	// 按请求时间排序
	sortEntriesByTime(harData.Log.Entries)

	// 计算序号位数
	totalEntries := len(harData.Log.Entries)
	digits := calculateDigits(totalEntries)

	report := &ExtractReport{
		TotalEntries: totalEntries,
		Results:      make([]ExtractResult, 0, totalEntries),
		Errors:       make([]ExtractError, 0),
	}

	for i := range harData.Log.Entries {
		entry := &harData.Log.Entries[i]
		index := i + 1 // 序号从1开始
		result := e.processEntry(entry, config, index, digits)
		report.Results = append(report.Results, result)

		if result.Success {
			report.SuccessCount++
		} else if result.Skipped {
			report.SkippedCount++
		} else {
			report.FailedCount++
			report.Errors = append(report.Errors, ExtractError{
				URL:   entry.Request.URL,
				Phase: "extract",
				Error: result.Error,
			})
		}

		// 根据策略决定是否继续
		if result.Error != nil && config.Strategy == StrategyStopOnFirstError {
			break
		}
	}

	return report, nil
}

func (e *extractor) ExtractParallel(harData *har.HAR, config *ExtractConfig) (*ExtractReport, error) {
	// 按请求时间排序
	sortEntriesByTime(harData.Log.Entries)

	// 计算序号位数
	totalEntries := len(harData.Log.Entries)
	digits := calculateDigits(totalEntries)

	workers := config.Workers
	if workers <= 0 {
		workers = 4 // 默认4个worker
	}

	// 使用带序号的任务channel
	taskChan := make(chan *IndexedTask, workers)
	resultsChan := make(chan ExtractResult, totalEntries)

	var wg sync.WaitGroup

	// 启动worker池
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				result := e.processEntry(task.Entry, config, task.Index, task.Digits)
				resultsChan <- result
			}
		}()
	}

	// 发送带序号的任务（序号在此处预分配，确保并发安全）
	go func() {
		for i := range harData.Log.Entries {
			task := &IndexedTask{
				Entry:  &harData.Log.Entries[i],
				Index:  i + 1, // 序号从1开始
				Digits: digits,
			}
			taskChan <- task
		}
		close(taskChan)
	}()

	// 等待完成并收集结果
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// 汇总结果
	report := &ExtractReport{
		TotalEntries: totalEntries,
		Results:      make([]ExtractResult, 0, totalEntries),
		Errors:       make([]ExtractError, 0),
	}

	for result := range resultsChan {
		report.Results = append(report.Results, result)
		if result.Success {
			report.SuccessCount++
		} else if result.Skipped {
			report.SkippedCount++
		} else {
			report.FailedCount++
			// 收集错误详情
			report.Errors = append(report.Errors, ExtractError{
				URL:   result.URL,
				Phase: "extract",
				Error: result.Error,
			})
		}
	}

	return report, nil
}

func (e *extractor) processEntry(entry *har.Entry, config *ExtractConfig, index int, digits int) ExtractResult {
	result := ExtractResult{
		URL:        entry.Request.URL,
		StatusCode: entry.Response.Status,
		MimeType:   entry.Response.Content.MimeType,
		Size:       entry.Response.Content.Size,
	}

	// 状态码过滤
	if len(config.FilterStatusCodes) > 0 {
		allowed := false
		for _, code := range config.FilterStatusCodes {
			if entry.Response.Status == code {
				allowed = true
				break
			}
		}
		if !allowed {
			result.Skipped = true
			return result
		}
	}

	// MIME类型过滤
	if len(config.FilterMimeTypes) > 0 {
		allowed := false
		for _, mime := range config.FilterMimeTypes {
			if e.mimeMapper.Match(entry.Response.Content.MimeType, mime) {
				allowed = true
				break
			}
		}
		if !allowed {
			result.Skipped = true
			return result
		}
	}

	// 检查空内容
	if config.Strategy == StrategySkipEmptyContent &&
		(entry.Response.Content.Text == "" || entry.Response.Content.Size == 0) {
		result.Skipped = true
		return result
	}

	// 解码内容
	data, err := e.decoder.Decode(&entry.Response.Content)
	if err != nil {
		result.Error = err
		return result
	}

	// 使用formatter格式化完整HTTP信息
	formattedOutput := e.formatter.FormatFullHTTP(entry, string(data))

	// 构建输出路径（带序号前缀）
	pathResult, err := e.pathBuilder.Build(entry.Request.URL, entry.Response.Content.MimeType, config.OutputDir, index, digits)
	if err != nil {
		result.Error = err
		return result
	}
	result.OutputPath = pathResult.ActualPath

	// 写入文件
	if err := e.writer.Write([]byte(formattedOutput), pathResult.ActualPath); err != nil {
		result.Error = err
		return result
	}

	result.Success = true
	return result
}
