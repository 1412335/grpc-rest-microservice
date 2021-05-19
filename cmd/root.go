package cmd

import (
	"os"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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
	// logger log.Logger
	// cmd
	rootCmd = &cobra.Command{
		Use:   "grpc-gateway",
		Short: "gRPC + gateway + tracing",
		Long:  `gRPC (gogo + validators) + gateway (openapi: swagger) + tracing (jaeger)`,
	}
)

func logError(logger log.Factory, err error) error {
	if err != nil {
		logger.Bg().Error("Error running cmd", zap.Error(err))
	}
	return err
}

func initConfig() {
	// load config from file
	cfgs = &configs.ServiceConfig{}
	if err := configs.LoadConfig(cfgFile, cfgs); err != nil {
		log.Fatal("Load config failed", zap.Error(err))
	}
	log.Info("Load config success", zap.String("file", viper.ConfigFileUsed()), zap.Any("config", cfgs))

	if cfgs.Log != nil {
		// set default logger
		log.DefaultLogger = log.NewFactory(log.WithLevel(cfgs.Log.Level), log.WithLevel(cfgs.Log.Level), log.WithTraceLevel(cfgs.Log.TraceLevel))
	}
	// // add serviceName + version to log
	// log.With(zap.String("service", cfgs.ServiceName), zap.String("version", cfgs.Version))
}

func init() {
	cobra.OnInitialize(initConfig)

	// cobra cmd bind args
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default $HOME/config.yml)")
	rootCmd.PersistentFlags().StringVarP(&version, "version", "v", "v1", "version")
	rootCmd.PersistentFlags().BoolVarP(&tracing, "tracing", "t", true, "using tracing with jaeger")
	rootCmd.PersistentFlags().StringVarP(&metricsBackend, "metrics", "m", "prometheus", "metrics backend expvar|prometheus")

	// bind from cobra cmd to viper
	// set from cmd args
	// cfgs.Tracing.Flag = tracing
	// cfgs.Tracing.Metrics = metricsBackend
	if err := viper.BindPFlag("version", rootCmd.PersistentFlags().Lookup("version")); err != nil {
		log.Error("Bind pflag version error", zap.Error(err))
	}
	if err := viper.BindPFlag("enableTracing", rootCmd.PersistentFlags().Lookup("tracing")); err != nil {
		log.Error("Bind pflag enableTracing error", zap.Error(err))
	}
	if err := viper.BindPFlag("tracing.metrics", rootCmd.PersistentFlags().Lookup("metrics")); err != nil {
		log.Error("Bind pflag tracing.metrics error", zap.Error(err))
	}

	// set logger
	// logger = log.DefaultLogger.Bg()
	log.Info("Root.Init")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal("Execute cmd failed", zap.Error(err))
		os.Exit(-1)
	}
}

func AddCommand(cmd ...*cobra.Command) {
	log.Info("Root.AddCommand")
	rootCmd.AddCommand(cmd...)
}

func LoadConfig() *configs.ServiceConfig {
	return cfgs
}
