package tracing

import (
	"fmt"
	"time"

	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/uber/jaeger-lib/metrics/expvar"
	"github.com/uber/jaeger-lib/metrics/prometheus"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/rpcmetrics"
	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"
)

var (
	defaultServiceName    = "default"
	defaultMetricsFactory = "prometheus"
	DefaultTracer         opentracing.Tracer
)

// Init creates a new instance of Jaeger tracer.
func Init(serviceName string, tracingMetrics string, logger log.Factory) opentracing.Tracer {
	cfg, err := config.FromEnv()
	if err != nil {
		logger.Bg().Fatal("cannot parse Jaeger env vars", zap.Error(err))
	}
	if serviceName == "" {
		serviceName = defaultServiceName
	}
	cfg.ServiceName = serviceName
	cfg.Sampler.Type = "const"
	cfg.Sampler.Param = 1

	logger.Bg().Info("jaeger config", zap.Any("config", cfg))

	// init metricsFactory
	var metricsFactory metrics.Factory
	if tracingMetrics == "expvar" {
		metricsFactory = expvar.NewFactory(10) // 10 buckets for histograms
		log.Info("[Tracing] Using expvar as metrics backend")
	} else {
		metricsFactory = prometheus.New().Namespace(metrics.NSOptions{Name: "tracing", Tags: nil})
		log.Info("[Tracing] Using prometheus as metrics backend")
	}

	// TODO(ys) a quick hack to ensure random generators get different seeds, which are based on current time.
	time.Sleep(100 * time.Millisecond)
	jaegerLogger := jaegerLoggerAdapter{logger.Bg()}

	metricsFactory = metricsFactory.Namespace(metrics.NSOptions{Name: serviceName, Tags: nil})
	tracer, _, err := cfg.NewTracer(
		config.Logger(jaegerLogger),
		config.Metrics(metricsFactory),
		config.Observer(rpcmetrics.NewObserver(metricsFactory, rpcmetrics.DefaultNameNormalizer)),
	)
	if err != nil {
		logger.Bg().Fatal("cannot initialize Jaeger Tracer", zap.Error(err))
	}
	return tracer
}

type jaegerLoggerAdapter struct {
	logger log.Logger
}

func (l jaegerLoggerAdapter) Error(msg string) {
	l.logger.Error(msg)
}

func (l jaegerLoggerAdapter) Infof(msg string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(msg, args...))
}
