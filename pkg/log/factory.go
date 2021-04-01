package log

import (
	"context"
	"os"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Factory struct {
	opts   *configs.Log
	logger *zap.Logger
}

func NewFactory(opts ...Option) Factory {
	// default options
	options := &configs.Log{
		Mode:  os.Getenv("GOENV"),
		Level: "FATAL",
	}

	f := Factory{
		opts: options,
	}
	// init logger w custom options
	if err := f.Init(opts...); err != nil {
		f.Bg().Error("config log options failed", zap.Error(err))
	}

	// new zap.Logger w mode
	var logger *zap.Logger
	if f.opts.Mode != "dev" {
		logger, _ = zap.NewProduction(
			zap.AddStacktrace(GetLevel(f.opts.Level)),
		)
	} else {
		logger, _ = zap.NewDevelopment(
			zap.AddStacktrace(GetLevel(f.opts.Level)),
			zap.AddCallerSkip(1),
		)
	}
	f.logger = logger
	return f
}

// override options
func (f *Factory) Init(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(f.opts); err != nil {
			return err
		}
	}
	return nil
}

func (f *Factory) Bg() Logger {
	return logger{logger: f.logger}
}

func (f *Factory) For(ctx context.Context) Logger {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		logger := spanLogger{span: span, logger: f.logger}
		if jaegerCtx, ok := span.Context().(jaeger.SpanContext); ok {
			logger.spanFields = []zapcore.Field{
				zap.String("trace_id", jaegerCtx.TraceID().String()),
				zap.String("span_id", jaegerCtx.SpanID().String()),
			}
		}
		return logger
	}
	return f.Bg()
}

func (f *Factory) With(fields ...zapcore.Field) Factory {
	return Factory{logger: f.logger.With(fields...)}
}
