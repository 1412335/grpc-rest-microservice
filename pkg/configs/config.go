package configs

import (
	"time"

	"github.com/spf13/viper"
)

const (
	CONFIG_NAME = "config"
	CONFIG_PATH = "."
	CONFIG_TYPE = "yml"
)

type ServiceConfig struct {
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
}

type GRPC struct {
	Host string
	Port int
}

// json web token
type JWT struct {
	SecretKey string
	Duration  time.Duration
}

type Database struct {
	Host     string
	User     string
	Password string
	Scheme   string
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

// grpc-gateway proxy
type Proxy struct {
	Port int
}

func LoadConfig() error {
	viper.SetConfigName(CONFIG_NAME)
	viper.SetConfigType(CONFIG_TYPE)
	viper.AddConfigPath(CONFIG_PATH)
	// Find and read the config file
	return viper.ReadInConfig()
}
