package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"har-decode/internal/extractor"
	"har-decode/internal/har"
	"har-decode/internal/logger"
	"har-decode/internal/output"
)

var (
	version = "1.0.0"
	commit  = "unknown"
	date    = "unknown"
)

// isDoubleClickMode 检测是否为双击运行模式（无命令行参数）
func isDoubleClickMode() bool {
	return len(os.Args) == 1
}

// getExeDir 获取程序exe所在目录
func getExeDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取程序路径失败: %w", err)
	}
	return filepath.Dir(exePath), nil
}

// findHarFiles 查找程序所在目录下的所有.har文件
func findHarFiles() ([]string, error) {
	// 获取程序所在目录
	exeDir, err := getExeDir()
	if err != nil {
		return nil, err
	}

	// 搜索程序所在目录下的.har文件
	pattern := filepath.Join(exeDir, "*.har")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("搜索HAR文件失败: %w", err)
	}

	return files, nil
}

// waitForExit 等待用户按键退出（防止窗口闪退）
func waitForExit() {
	fmt.Println("\n按 Enter 键退出...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

// runAutoMode 自动模式：搜索并处理当前目录下的.har文件
func runAutoMode() error {
	fmt.Println("========================================")
	fmt.Printf("  HAR文件分离工具 v%s\n", version)
	fmt.Println("========================================")
	fmt.Println("\n正在搜索当前目录下的HAR文件...")

	// 查找.har文件
	harFiles, err := findHarFiles()
	if err != nil {
		return err
	}

	if len(harFiles) == 0 {
		fmt.Println("\n未找到任何.har文件！")
		fmt.Println("请将本程序放在包含.har文件的目录中运行。")
		return nil
	}

	fmt.Printf("\n找到 %d 个HAR文件:\n", len(harFiles))
	for i, f := range harFiles {
		fmt.Printf("  %d. %s\n", i+1, filepath.Base(f))
	}

	// 初始化日志
	log, err := logger.NewZapLogger(logger.InfoLevel, false)
	if err != nil {
		return fmt.Errorf("初始化日志失败: %w", err)
	}

	// 获取程序所在目录
	exeDir, err := getExeDir()
	if err != nil {
		return err
	}

	// 处理每个HAR文件
	successCount := 0
	for i, harFile := range harFiles {
		fmt.Printf("\n[%d/%d] 正在处理: %s\n", i+1, len(harFiles), filepath.Base(harFile))

		// 根据HAR文件名构建输出目录: {程序目录}/decode/{har文件名}/
		harBaseName := strings.TrimSuffix(filepath.Base(harFile), ".har")
		outputDir := filepath.Join(exeDir, "decode", harBaseName)

		// 默认配置
		config := &extractor.ExtractConfig{
			OutputDir:         outputDir,
			Strategy:          extractor.StrategyContinueOnError,
			Workers:           4,
			FilterMimeTypes:   nil,
			FilterStatusCodes: nil,
			SkipMediaFiles:    true, // 双击模式默认跳过图片和CSS
			Verbose:           false,
		}

		if err := runExtract(harFile, config, log); err != nil {
			fmt.Printf("  ✗ 处理失败: %v\n", err)
			continue
		}
		successCount++
	}

	fmt.Println("\n========================================")
	fmt.Printf("处理完成！成功: %d/%d\n", successCount, len(harFiles))
	outputDir := filepath.Join(exeDir, "decode")
	fmt.Printf("输出目录: %s\n", outputDir)
	fmt.Println("========================================")

	return nil
}

func main() {
	// 检测双击运行模式
	clickMode := isDoubleClickMode()

	if clickMode {
		// 双击模式：自动处理
		if err := runAutoMode(); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		}
		waitForExit()
		os.Exit(0)
	}

	// 命令行模式：正常执行
	if err := execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func execute() error {
	var (
		inputFile       string
		outputDir       string
		verbose         bool
		filterMime      string
		filterStatus    string
		workers         int
		continueOnError bool
		stopOnError     bool
		skipEmpty       bool
		showVersion     bool
		parseAll        bool // 解析所有内容（包括图片和CSS）
	)

	rootCmd := &cobra.Command{
		Use:   "har-decode",
		Short: "HAR文件分离工具 - 将HAR文件中的资源提取为独立文件",
		Long: `HAR文件分离工具用于将浏览器导出的HAR（HTTP Archive）文件分离为独立的原始文件。

支持功能：
  - 流式解析大文件
  - 并发处理提高效率
  - MIME类型和状态码过滤
  - 自动处理文件名冲突
  - 支持base64编码内容解码`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				fmt.Printf("har-decode version %s (commit: %s, built: %s)\n", version, commit, date)
				return nil
			}

			if inputFile == "" {
				return fmt.Errorf("必须指定输入文件，使用 --input 或 -i 参数")
			}

			// 初始化日志
			logLevel := logger.InfoLevel
			if verbose {
				logLevel = logger.DebugLevel
			}
			log, err := logger.NewZapLogger(logLevel, verbose)
			if err != nil {
				return fmt.Errorf("初始化日志失败: %w", err)
			}

			// 解析过滤条件
			var mimeTypes []string
			if filterMime != "" {
				mimeTypes = strings.Split(filterMime, ",")
				for i, m := range mimeTypes {
					mimeTypes[i] = strings.TrimSpace(m)
				}
			}

			var statusCodes []int
			if filterStatus != "" {
				codes := strings.Split(filterStatus, ",")
				for _, c := range codes {
					c = strings.TrimSpace(c)
					var code int
					if _, err := fmt.Sscanf(c, "%d", &code); err == nil {
						statusCodes = append(statusCodes, code)
					}
				}
			}

			// 确定策略
			strategy := extractor.StrategyContinueOnError
			if stopOnError {
				strategy = extractor.StrategyStopOnFirstError
			}
			if skipEmpty {
				strategy = extractor.StrategySkipEmptyContent
			}

			// 如果未指定输出目录，使用默认目录: {程序目录}/decode/{har文件名}/
			if outputDir == "" {
				exeDir, err := getExeDir()
				if err != nil {
					return err
				}
				harBaseName := strings.TrimSuffix(filepath.Base(inputFile), ".har")
				outputDir = filepath.Join(exeDir, "decode", harBaseName)
			}

			// 构建配置
			config := &extractor.ExtractConfig{
				OutputDir:         outputDir,
				Strategy:          strategy,
				Workers:           workers,
				FilterMimeTypes:   mimeTypes,
				FilterStatusCodes: statusCodes,
				SkipMediaFiles:    !parseAll, // 默认跳过图片和CSS，-a参数时设为false
				Verbose:           verbose,
			}

			// 执行提取
			return runExtract(inputFile, config, log)
		},
	}

	// 命令行参数
	rootCmd.Flags().StringVarP(&inputFile, "input", "i", "", "HAR文件路径（必填）")
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "", "输出目录（默认: {程序目录}/decode/{har文件名}/）")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "显示详细输出")
	rootCmd.Flags().StringVarP(&filterMime, "filter", "f", "", "MIME类型过滤（逗号分隔，支持*通配符）")
	rootCmd.Flags().StringVarP(&filterStatus, "status", "s", "", "HTTP状态码过滤（逗号分隔）")
	rootCmd.Flags().IntVarP(&workers, "workers", "w", 4, "并发worker数量")
	rootCmd.Flags().BoolVarP(&continueOnError, "continue-on-error", "c", true, "遇到错误继续处理")
	rootCmd.Flags().BoolVar(&stopOnError, "stop-on-error", false, "遇到第一个错误停止")
	rootCmd.Flags().BoolVar(&skipEmpty, "skip-empty", false, "跳过空内容条目")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "V", false, "显示版本信息")
	rootCmd.Flags().BoolVarP(&parseAll, "all", "a", false, "解析所有内容（包括图片和CSS文件）")

	// 绑定viper
	viper.BindPFlag("output", rootCmd.Flags().Lookup("output"))
	viper.BindPFlag("verbose", rootCmd.Flags().Lookup("verbose"))
	viper.BindPFlag("workers", rootCmd.Flags().Lookup("workers"))

	// 配置文件支持
	viper.SetConfigName(".har-decode")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")
	viper.AutomaticEnv()

	_ = viper.ReadInConfig() // 忽略配置文件不存在的错误

	return rootCmd.Execute()
}

func runExtract(inputFile string, config *extractor.ExtractConfig, log logger.Logger) error {
	log.Info("开始处理HAR文件",
		logger.F("input", inputFile),
		logger.F("output", config.OutputDir))

	// 创建解析器
	parser := har.NewParser(log)

	// 解析HAR文件
	harData, err := parser.Parse(inputFile)
	if err != nil {
		return fmt.Errorf("解析HAR文件失败: %w", err)
	}

	log.Info("HAR文件解析完成",
		logger.F("entries", len(harData.Log.Entries)))

	// 创建组件
	writer := output.NewWriter()
	decoder := extractor.NewContentDecoder()
	mimeMapper := extractor.NewMimeTypeMapper()
	conflictResolver := output.NewConflictResolver()
	pathBuilder := output.NewPathBuilder(mimeMapper, conflictResolver, log)

	// 创建提取器
	ex := extractor.NewExtractor(writer, decoder, pathBuilder, mimeMapper, log)

	// 执行提取
	var report *extractor.ExtractReport
	if config.Workers > 1 {
		report, err = ex.ExtractParallel(harData, config)
	} else {
		report, err = ex.Extract(harData, config)
	}

	if err != nil {
		return fmt.Errorf("提取文件失败: %w", err)
	}

	// 输出报告
	printReport(report, config.Verbose)

	return nil
}

func printReport(report *extractor.ExtractReport, verbose bool) {
	fmt.Println("\n========== 提取报告 ==========")
	fmt.Printf("总条目数: %d\n", report.TotalEntries)
	fmt.Printf("成功提取: %d\n", report.SuccessCount)
	fmt.Printf("跳过条目: %d\n", report.SkippedCount)
	fmt.Printf("失败条目: %d\n", report.FailedCount)

	if len(report.Errors) > 0 {
		fmt.Println("\n错误列表:")
		for i, e := range report.Errors {
			if i >= 10 {
				fmt.Printf("... 还有 %d 个错误\n", len(report.Errors)-10)
				break
			}
			fmt.Printf("  - %s: %v\n", e.URL, e.Error)
		}
	}

	if verbose && len(report.Results) > 0 {
		fmt.Println("\n提取详情:")
		for _, r := range report.Results {
			if r.Success {
				fmt.Printf("  ✓ %s -> %s\n", r.URL, r.OutputPath)
			} else if r.Skipped {
				fmt.Printf("  - %s (跳过)\n", r.URL)
			} else {
				fmt.Printf("  ✗ %s (失败: %v)\n", r.URL, r.Error)
			}
		}
	}

	fmt.Println("==============================")
}
