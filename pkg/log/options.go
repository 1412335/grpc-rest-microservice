package log

import (
	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"go.uber.org/zap/zapcore"
)

type Option func(*configs.Log) error

func WithLevel(level string) Option {
	return func(o *configs.Log) error {
		if level != "" {
			o.Level = level
		}
		return nil
	}
}

func GetLevel(level string) zapcore.Level {
	switch level {
	case "PANNIC":
		return zapcore.PanicLevel
	case "FATAL":
		return zapcore.FatalLevel
	case "ERROR":
		return zapcore.ErrorLevel
	case "WARN":
		return zapcore.WarnLevel
	case "INFO":
		return zapcore.InfoLevel
	case "DEBUG":
		return zapcore.DebugLevel
	default:
		return zapcore.InfoLevel
	}
}
