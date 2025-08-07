package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.SugaredLogger
}

type Config struct {
	Level      string
	Format     string
	OutputPath string
}

func NewLogger(config Config) (*Logger, error) {
	// Parse log level
	var level zapcore.Level
	switch config.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	// Configure encoder
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if config.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Configure output
	var writeSyncer zapcore.WriteSyncer
	if config.OutputPath == "stdout" || config.OutputPath == "" {
		writeSyncer = zapcore.Lock(os.Stdout)
	} else {
		file, err := os.OpenFile(config.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		writeSyncer = zapcore.Lock(file)
	}

	// Create core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// Create logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{
		SugaredLogger: zapLogger.Sugar(),
	}, nil
}

func (l *Logger) WithTenant(tenantID string) *Logger {
	return &Logger{
		SugaredLogger: l.SugaredLogger.With("tenant_id", tenantID),
	}
}

func (l *Logger) WithUser(userID string) *Logger {
	return &Logger{
		SugaredLogger: l.SugaredLogger.With("user_id", userID),
	}
}

func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		SugaredLogger: l.SugaredLogger.With("request_id", requestID),
	}
}

func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{
		SugaredLogger: l.SugaredLogger.With(args...),
	}
}

// Global logger instance
var globalLogger *Logger

func Init(config Config) error {
	logger, err := NewLogger(config)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

func Debug(args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(args...)
	}
}

func Info(args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Info(args...)
	}
}

func Warn(args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(args...)
	}
}

func Error(args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Error(args...)
	}
}

func Fatal(args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Fatal(args...)
	}
}

func Debugf(template string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Debugf(template, args...)
	}
}

func Infof(template string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Infof(template, args...)
	}
}

func Warnf(template string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Warnf(template, args...)
	}
}

func Errorf(template string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Errorf(template, args...)
	}
}

func Fatalf(template string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Fatalf(template, args...)
	}
}

func GetLogger() *Logger {
	return globalLogger
}
