package cmd

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"time"

	"grpc-rest-microservice/pkg/protocol/grpc"
	v1 "grpc-rest-microservice/pkg/service/v1"
	v2 "grpc-rest-microservice/pkg/service/v2"
	"grpc-rest-microservice/pkg/utils"
)

type Config struct {
	// gRPC is TCP port to listen by gRPC server
	GRPCPort string

	JWTSecretKey string
	JWTDuration  int

	DBHost     string
	DBUser     string
	DBPassword string
	DBScheme   string
}

func RunServer() error {
	ctx := context.Background()

	var config Config
	flag.StringVar(&config.GRPCPort, "grpc-port", "", "gRPC port to bind")
	flag.StringVar(&config.DBHost, "db-host", "", "Database host")
	flag.StringVar(&config.DBUser, "db-user", "", "Database user")
	flag.StringVar(&config.DBPassword, "db-password", "", "Database pwd")
	flag.StringVar(&config.DBScheme, "db-scheme", "", "Database scheme")
	flag.Parse()

	if len(config.GRPCPort) == 0 {
		return fmt.Errorf("invalid TCP port for gRPC server: %s", config.GRPCPort)
	}

	param := "parseTime=true"
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s",
		config.DBUser,
		config.DBPassword,
		config.DBHost,
		config.DBScheme,
		param)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect database: %v", err)
	}
	defer db.Close()

	v1API := v1.NewToDoServiceServer(db)

	return grpc.RunServer(ctx, v1API, config.GRPCPort)
}

func RunServerV2() error {
	ctx := context.Background()

	var config Config
	flag.StringVar(&config.GRPCPort, "grpc-port", "", "gRPC port to bind")
	flag.StringVar(&config.JWTSecretKey, "jwt-secret", "luloveyen", "jwt secret key")
	flag.IntVar(&config.JWTDuration, "jwt-duration", 60, "jwt token duration (seconds)")
	flag.Parse()

	if len(config.GRPCPort) == 0 {
		return fmt.Errorf("invalid TCP port for gRPC server v2: %s", config.GRPCPort)
	}

	v2API := v2.NewServiceAImpl()

	// create implementation extra service
	secretKey := config.JWTSecretKey
	tokenDuration := time.Duration(config.JWTDuration) * time.Second
	jwtManager := utils.NewJWTManager(secretKey, tokenDuration)

	v2API_extra := v2.NewServiceExtraImpl(jwtManager)

	// server interceptor
	// serverInterceptor := grpc.SimpleServerInterceptor{}

	const v2ServicePath = "/v2.ServiceExtra/"
	accessibleRoles := map[string][]string{
		v2ServicePath + "Post": {"admin"},
	}
	serverInterceptor := grpc.NewAuthServerInterceptor(jwtManager, accessibleRoles)

	return grpc.RunServerV2(ctx, serverInterceptor, v2API, v2API_extra, config.GRPCPort)
}
