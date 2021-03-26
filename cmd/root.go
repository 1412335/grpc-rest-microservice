package cmd

import (
	"os"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-lib/metrics"
	"github.com/uber/jaeger-lib/metrics/expvar"
	jprom "github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Used for flags.
	cfgFile        string
	version        string
	tracing        bool
	metricsBackend string
	// service config
	cfgs *configs.ServiceConfig
	// log
	logger *zap.Logger
	// tracing
	metricsFactory metrics.Factory
	// cmd
	rootCmd = &cobra.Command{
		Use:   "grpc-gateway",
		Short: "gRPC + gateway + tracing",
		Long:  `gRPC (gogo + validators) + gateway (openapi: swagger) + tracing (jaeger)`,
	}
)

func logError(logger *zap.Logger, err error) error {
	if err != nil {
		logger.Error("Error running cmd", zap.Error(err))
	}
	return err
}

func initConfig() {
	// load config from file
	cfgs = &configs.ServiceConfig{}
	if err := configs.LoadConfig(cfgFile, cfgs); err != nil {
		logger.Fatal("Load config failed", zap.Error(err))
	}

	logger.Info("Load config success", zap.String("file", viper.ConfigFileUsed()), zap.Any("config", cfgs))
	// set from cmd args
	// cfgs.Tracing.Flag = tracing
	// cfgs.Tracing.Metrics = metricsBackend

	// tracing
	if cfgs.EnableTracing {
		if cfgs.Tracing != nil && cfgs.Tracing.Metrics == "expvar" {
			metricsFactory = expvar.NewFactory(10) // 10 buckets for histograms
			logger.Info("[Tracing] Using expvar as metrics backend")
		} else {
			metricsFactory = jprom.New().Namespace(metrics.NSOptions{Name: "tracing", Tags: nil})
			logger.Info("[Tracing] Using prometheus as metrics backend")
		}
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// cobra cmd bind args
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default $HOME/config.yml)")
	rootCmd.PersistentFlags().StringVarP(&version, "version", "v", "v1", "version")
	rootCmd.PersistentFlags().BoolVarP(&tracing, "tracing", "t", true, "using tracing with jaeger")
	rootCmd.PersistentFlags().StringVarP(&metricsBackend, "metrics", "m", "prometheus", "metrics backend expvar|prometheus")

	// bind from cobra cmd to viper
	if err := viper.BindPFlag("version", rootCmd.PersistentFlags().Lookup("version")); err != nil {
		logger.Error("Bind pflag version error", zap.Error(err))
	}
	if err := viper.BindPFlag("tracing.metrics", rootCmd.PersistentFlags().Lookup("metrics")); err != nil {
		logger.Error("Bind pflag tracing.metrics error", zap.Error(err))
	}
	if err := viper.BindPFlag("tracing.flag", rootCmd.PersistentFlags().Lookup("tracing")); err != nil {
		logger.Error("Bind pflag tracing.flag error", zap.Error(err))
	}

	// set logger
	logger, _ = zap.NewDevelopment(
		zap.AddStacktrace(zapcore.FatalLevel),
		zap.AddCallerSkip(1),
	)
	logger.Info("Root.Init")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal("Execute cmd failed", zap.Error(err))
		os.Exit(-1)
	}
}
