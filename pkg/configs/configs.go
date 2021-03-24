package configs

import (
	"time"

	"github.com/spf13/viper"
)

const (
	ConfigName = "config"
	ConfigType = "yml"
	ConfigPath = "."
)

type ServiceConfig struct {
	Version        string
	ServiceName    string
	GRPC           *GRPC
	Proxy          *Proxy
	ManagerClient  *ManagerClient
	JWT            *JWT
	Database       *Database
	Authentication *Authentication
	// server using
	AccessibleRoles map[string][]string
	// client using
	AuthMethods map[string]bool
	// opentracing
	Tracing *Tracing
	// swaggers
	Swagger []string
}

type ClientConfig struct {
	Version       string
	ServiceName   string
	GRPC          *GRPC
	ManagerClient *ManagerClient
}

// opentracing with jaeger
type Tracing struct {
	Flag    bool
	Metrics string
}

// grpc-server
type GRPC struct {
	Host string
	Port int
}

// grpc-gateway proxy
type Proxy struct {
	Port int
}

// json web token
type JWT struct {
	Issuer    string
	SecretKey string
	Duration  time.Duration
}

// db
// type Database struct {
// 	Name    string
// 	MySQL   *MySQL
// 	MongoDB *MongoDB
// }

// mysql
type Database struct {
	Host           string
	Port           string
	User           string
	Password       string
	Scheme         string
	Debug          bool
	MaxIdleConns   int
	MaxOpenConns   int
	ConnectTimeout time.Duration
}

// mongodb
type MongoDB struct {
	ConnectionURI   string
	Database        string
	Auth            *MongoDBAuthentication
	PoolSize        uint64
	MaxConnIdleTime time.Duration
	ConnectTimeout  time.Duration
	Debug           bool
}

type MongoDBAuthentication struct {
	Source   string
	User     string
	Password string
}

// manager grpc-pool
type ManagerClient struct {
	MaxPoolSize int
	TimeOut     int
	// method need to request with authentication
	AuthMethods map[string]bool
	// credentials authentication
	Authentication *Authentication
	// jwt token
	RefreshDuration time.Duration
}

type Authentication struct {
	Username string
	Password string
}

// type AccessibleRoles map[string][]string

func LoadConfig(cfgFile string, cfg interface{}) error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".cobra" (without extension).
		viper.SetConfigName(ConfigName)
		viper.SetConfigType(ConfigType)
		viper.AddConfigPath(ConfigPath)
	}
	viper.AutomaticEnv()
	// Find and read the config file
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.Unmarshal(cfg); err != nil {
		return err
	}
	return nil
}
