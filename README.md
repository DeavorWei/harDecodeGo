# HAR文件分离工具

一个高效的Golang命令行工具，用于将浏览器导出的HAR（HTTP Archive）文件分离为独立的原始文件。

## 功能特性

- **HAR文件解析**：解析JSON格式的HAR文件，支持流式解析处理大文件
- **内容提取**：从每个entry中提取响应内容，支持并发处理
- **文件保存**：根据URL路径结构保存文件，处理文件名冲突
- **编码处理**：支持base64编码的二进制内容解码
- **冲突处理**：处理同名文件的命名冲突，支持路径长度限制
- **过滤功能**：支持MIME类型和HTTP状态码过滤

## 安装

### 从源码编译

```bash
# Windows
build.bat

# 或手动编译
go build -o bin/har-decode.exe ./cmd/har-decode
```

## 使用方法

### 基本用法

```bash
# 基本用法
har-decode --input <har文件路径> --output <输出目录>

# 示例
har-decode --input ./TestData/example.har --output ./output

# 显示详细输出
har-decode --input ./TestData/example.har --output ./output --verbose

# 过滤特定MIME类型（支持通配符）
har-decode --input ./TestData/example.har --output ./output --filter "image/*,application/javascript"

# 过滤多个状态码（逗号分隔）
har-decode --input ./TestData/example.har --output ./output --status "200,301,302"

# 并发处理（指定worker数量）
har-decode --input ./TestData/example.har --output ./output --workers 8

# 跳过空内容条目
har-decode --input ./TestData/example.har --output ./output --skip-empty

# 显示帮助
har-decode --help
```

### 命令行参数

| 长参数              | 简写 | 说明                                   | 默认值   | 是否必填 |
| ------------------- | ---- | -------------------------------------- | -------- | -------- |
| --input             | -i   | HAR文件路径                            | -        | 是       |
| --output            | -o   | 输出目录                               | ./output | 否       |
| --verbose           | -v   | 显示详细输出                           | false    | 否       |
| --filter            | -f   | MIME类型过滤（逗号分隔，支持\*通配符） | 全部     | 否       |
| --status            | -s   | HTTP状态码过滤（逗号分隔）             | 全部     | 否       |
| --workers           | -w   | 并发worker数量                         | 4        | 否       |
| --continue-on-error | -c   | 遇到错误继续处理                       | true     | 否       |
| --stop-on-error     | -    | 遇到第一个错误停止                     | false    | 否       |
| --skip-empty        | -    | 跳过空内容条目                         | false    | 否       |
| --help              | -h   | 显示帮助信息                           | -        | -        |
| --version           | -V   | 显示版本信息                           | -        | -        |

### 配置文件支持

支持从配置文件读取默认参数（可选）：

```yaml
# .har-decode.yaml
output: ./output
workers: 4
verbose: false
skip_empty: true
filter:
  - "image/*"
  - "application/javascript"
status:
  - 200
  - 304
```

配置文件查找顺序：

1. 当前目录 `.har-decode.yaml`
2. 用户主目录 `~/.har-decode.yaml`

## 项目结构

```
har-decode/
├── cmd/
│   └── har-decode/
│       └── main.go              # 程序入口
├── internal/
│   ├── har/
│   │   ├── har.go               # HAR数据结构定义
│   │   ├── parser.go            # HAR解析器（支持流式解析）
│   │   └── error.go             # 错误定义
│   ├── extractor/
│   │   ├── extractor.go         # 文件提取器（支持并发）
│   │   ├── encoding.go          # 编码处理
│   │   └── mime.go              # MIME类型映射
│   ├── output/
│   │   ├── writer.go            # 文件写入器
│   │   ├── path_builder.go      # 路径构建器
│   │   └── conflict.go          # 文件名冲突处理
│   └── logger/
│       ├── logger.go            # 日志接口
│       └── zap_logger.go        # zap实现
├── pkg/
│   └── utils/
│       ├── url.go               # URL工具函数
│       ├── file.go              # 文件工具函数
│       └── sanitize.go          # 文件名清理
├── TestData/                    # 测试数据
├── docs/                        # 文档
├── go.mod
├── go.sum
├── build.bat                    # 编译脚本
└── README.md
```

## 开发

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行并生成覆盖率报告
go test -cover ./...

# 运行基准测试
go test -bench=. ./...
```

### 依赖

- [github.com/spf13/cobra](https://github.com/spf13/cobra) - 命令行框架
- [github.com/spf13/viper](https://github.com/spf13/viper) - 配置管理
- [go.uber.org/zap](https://go.uber.org/zap) - 高性能日志

## 许可证

MIT License
