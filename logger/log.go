package logger

import (
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

const (
	_defaultLogFile = "log.log"
)

var _globalLogger *ZapLogger // caller skip 1

type LoggerConfig struct {
	Level      string `yaml:"level"`       // debug  info  warn  error
	Encoding   string `yaml:"encoding"`    // json or console
	CallFull   bool   `yaml:"call_full"`   // whether full call path or short path, default is short
	Filename   string `yaml:"file_name"`   // log file name
	MaxSize    int    `yaml:"max_size"`    // max size of log.(MB)
	MaxAge     int    `yaml:"max_age"`     // time to keep, (day)
	MaxBackups int    `yaml:"max_backups"` // max file numbers
	LocalTime  bool   `yaml:"local_time"`  //(default UTC)
	Compress   bool   `yaml:"compress"`    // default false
	CallerSkip int    `yaml:"caller_skip"` // 可选项
}

var d = `level: info
max_size: 50
max_age: 1
env: prod
max_backups: 2
is_test: 0
encoding : console`

func init() {
	tmp := []byte(d)
	loggerConfig := &LoggerConfig{}
	if err := yaml.Unmarshal(tmp, loggerConfig); err != nil {
		panic(err)
	}
	initGlobalLogger(loggerConfig)
}

func initGlobalLogger(loggerConfig *LoggerConfig) {
	if loggerConfig.Filename == "" {
		loggerConfig.Filename = _defaultLogFile
	}

	enCfg := zap.NewProductionEncoderConfig()
	enCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	if loggerConfig.CallFull {
		enCfg.EncodeCaller = zapcore.FullCallerEncoder
	}
	encoder := zapcore.NewJSONEncoder(enCfg)

	zapWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   loggerConfig.Filename,
		MaxSize:    loggerConfig.MaxSize,
		MaxAge:     loggerConfig.MaxAge,
		MaxBackups: loggerConfig.MaxBackups,
		LocalTime:  loggerConfig.LocalTime,
	})
	if loggerConfig.Encoding == "console" {
		zapWriter = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(zapWriter))
	}

	newCore := zapcore.NewCore(encoder, zapWriter, zap.NewAtomicLevelAt(convertLogLevel(loggerConfig.Level)))
	opts := []zap.Option{
		zap.ErrorOutput(zapWriter),
		zap.AddCaller(),
		// zap.AddStacktrace(zapcore.WarnLevel),
		zap.AddCallerSkip(2),
	}

	_globalLogger = NewZapLogger(zap.New(newCore, opts...))
}

func convertLogLevel(levelStr string) (level zapcore.Level) {
	switch levelStr {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	}
	return
}
