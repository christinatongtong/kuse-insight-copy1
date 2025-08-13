package logger

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const FIELDS = "fields"

var LoggerEncoder zapcore.Encoder

type ZapLogger struct {
	logger *zap.Logger
}

func init() {
	enCfg := zap.NewProductionEncoderConfig()
	enCfg.LevelKey = ""
	enCfg.TimeKey = ""
	enCfg.CallerKey = ""
	enCfg.MessageKey = ""
	enCfg.NameKey = ""
	enCfg.FunctionKey = ""
	enCfg.StacktraceKey = ""
	LoggerEncoder = zapcore.NewJSONEncoder(enCfg)
}

func NewZapLogger(logger *zap.Logger) *ZapLogger {
	return &ZapLogger{
		logger: logger,
	}
}

func transform2ZapField(logger *zap.Logger, level zapcore.Level, msg string, fields ...zapcore.Field) zap.Field {
	filterdFields := make([]zapcore.Field, 0, len(fields))
	for i := range fields {
		if fields[i].Key != "" {
			if fields[i].Type == zapcore.ObjectMarshalerType || fields[i].Type == zapcore.ReflectType {
				filterdFields = append(filterdFields, zap.String(fields[i].Key, fmt.Sprintf("%+v", fields[i].Interface)))
			} else {
				filterdFields = append(filterdFields, fields[i])
			}
		}
	}

	var raw json.RawMessage
	if ce := logger.Check(level, msg); ce != nil {
		buf, _ := LoggerEncoder.EncodeEntry(ce.Entry, filterdFields)
		raw = buf.Bytes()
	}

	return zap.Any(FIELDS, &raw)
}

func (zl *ZapLogger) With(fields ...zapcore.Field) *ZapLogger {
	return NewZapLogger(zl.logger.With(fields...))
}

func (zl *ZapLogger) WithOptions(opts ...zap.Option) *ZapLogger {
	return NewZapLogger(zl.logger.WithOptions(opts...))
}

func (zl *ZapLogger) Printf(format string, v ...interface{}) {
	zl.logger.Info(fmt.Sprintf(format, v...))
}

func (zl *ZapLogger) Info(msg string, fields ...zapcore.Field) {
	zl.logger.Info(msg, transform2ZapField(zl.logger, zap.InfoLevel, msg, fields...))
}

func (zl *ZapLogger) Debug(msg string, fields ...zapcore.Field) {
	zl.logger.Debug(msg, transform2ZapField(zl.logger, zap.DebugLevel, msg, fields...))
}

func (zl *ZapLogger) Warn(msg string, fields ...zapcore.Field) {
	zl.logger.Warn(msg, transform2ZapField(zl.logger, zap.WarnLevel, msg, fields...))
}

func (zl *ZapLogger) Error(msg string, fields ...zapcore.Field) {
	zl.logger.Error(msg, transform2ZapField(zl.logger, zap.ErrorLevel, msg, fields...))
}

func (zl *ZapLogger) Fatal(msg string, fields ...zapcore.Field) {
	zl.logger.Fatal(msg, transform2ZapField(zl.logger, zap.FatalLevel, msg, fields...))
}

func (zl *ZapLogger) Panic(msg string, fields ...zapcore.Field) {
	zl.logger.Panic(msg, transform2ZapField(zl.logger, zap.PanicLevel, msg, fields...))
}

// 使用原生zap打印，单个字段会被索引为ES中的一列
func (zl *ZapLogger) InfoI(msg string, fields ...zapcore.Field) {
	zl.logger.Info(msg, fields...)
}

// 使用原生zap打印，单个字段会被索引为ES中的一列
func (zl *ZapLogger) DebugI(msg string, fields ...zapcore.Field) {
	zl.logger.Debug(msg, fields...)
}

// 使用原生zap打印，单个字段会被索引为ES中的一列
func (zl *ZapLogger) WarnI(msg string, fields ...zapcore.Field) {
	zl.logger.Warn(msg, fields...)
}

// 使用原生zap打印，单个字段会被索引为ES中的一列
func (zl *ZapLogger) ErrorI(msg string, fields ...zapcore.Field) {
	zl.logger.Error(msg, fields...)
}

// 使用原生zap打印，单个字段会被索引为ES中的一列
func (zl *ZapLogger) FatalI(msg string, fields ...zapcore.Field) {
	zl.logger.Fatal(msg, fields...)
}

// 使用原生zap打印，单个字段会被索引为ES中的一列
func (zl *ZapLogger) PanicI(msg string, fields ...zapcore.Field) {
	zl.logger.Panic(msg, fields...)
}
