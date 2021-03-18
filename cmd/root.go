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
		Short: "gRPC + gateway (openapi: swagger) + tracing (jaeger)",
		Long:  `gRPC + gateway (openapi: swagger) + tracing (jaeger)`,
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

	// set from cmd args
	cfgs.Tracing.Flag = tracing
	cfgs.Tracing.Metrics = metricsBackend

	logger.Info("Load config success", zap.String("file", viper.ConfigFileUsed()))
	// tracing
	if cfgs.Tracing.Flag {
		if cfgs.Tracing.Metrics == "expvar" {
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
	rootCmd.PersistentFlags().BoolVarP(&tracing, "tracing", "t", true, "using tracing with jaeger")
	rootCmd.PersistentFlags().StringVarP(&metricsBackend, "metrics", "m", "prometheus", "metrics backend expvar|prometheus")

	// bind from viper to cobra cmd
	viper.BindPFlag("metrics", rootCmd.PersistentFlags().Lookup("metrics"))
	viper.BindPFlag("tracing", rootCmd.PersistentFlags().Lookup("tracing"))

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
