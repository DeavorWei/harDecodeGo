package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger zap日志实现
type ZapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
	fields []Field
}

// NewZapLogger 创建zap日志实例
func NewZapLogger(level Level, verbose bool) (Logger, error) {
	config := zap.NewProductionConfig()

	if verbose {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// 设置日志级别
	switch level {
	case DebugLevel:
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case InfoLevel:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case WarnLevel:
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case ErrorLevel:
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	case FatalLevel:
		config.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	}

	zapLogger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &ZapLogger{
		logger: zapLogger,
		sugar:  zapLogger.Sugar(),
	}, nil
}

// NewNopLogger 创建空操作日志（用于测试）
func NewNopLogger() Logger {
	return &ZapLogger{
		logger: zap.NewNop(),
		sugar:  zap.NewNop().Sugar(),
	}
}

func (l *ZapLogger) convertFields(fields ...Field) []interface{} {
	result := make([]interface{}, 0, len(fields)*2)
	for _, f := range fields {
		result = append(result, f.Key, f.Value)
	}
	// 添加预设字段
	for _, f := range l.fields {
		result = append(result, f.Key, f.Value)
	}
	return result
}

func (l *ZapLogger) Debug(msg string, fields ...Field) {
	l.sugar.Debugw(msg, l.convertFields(fields...)...)
}

func (l *ZapLogger) Info(msg string, fields ...Field) {
	l.sugar.Infow(msg, l.convertFields(fields...)...)
}

func (l *ZapLogger) Warn(msg string, fields ...Field) {
	l.sugar.Warnw(msg, l.convertFields(fields...)...)
}

func (l *ZapLogger) Error(msg string, fields ...Field) {
	l.sugar.Errorw(msg, l.convertFields(fields...)...)
}

func (l *ZapLogger) Fatal(msg string, fields ...Field) {
	l.sugar.Fatalw(msg, l.convertFields(fields...)...)
}

func (l *ZapLogger) With(fields ...Field) Logger {
	return &ZapLogger{
		logger: l.logger,
		sugar:  l.sugar,
		fields: append(l.fields, fields...),
	}
}

// Sync 刷新日志缓冲
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}
