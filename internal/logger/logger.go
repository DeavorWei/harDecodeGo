package logger

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// Logger 日志接口
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
}

// Level 日志级别
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// F 创建日志字段的快捷方法
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}
