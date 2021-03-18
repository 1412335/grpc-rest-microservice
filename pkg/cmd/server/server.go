package cmd

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/interceptor"
	"github.com/1412335/grpc-rest-microservice/pkg/protocol/grpc"
	v1 "github.com/1412335/grpc-rest-microservice/pkg/service/v1"
	v2 "github.com/1412335/grpc-rest-microservice/pkg/service/v2"
	"github.com/1412335/grpc-rest-microservice/pkg/utils"
)

func RunServer() error {
	srvConfig := &configs.ServiceConfig{}
	if err := configs.LoadConfig("", srvConfig); err != nil {
		log.Fatalf("[Main] Load config failed: %v", err)
	}
	// if err := viper.Unmarshal(srvConfig); err != nil {
	// 	log.Fatalf("[Main] Unmarshal config failed: %v", err)
	// }

	GRPCPort := flag.Int("grpc-port", 9090, "gRPC port service")
	DBHost := flag.String("db-host", "", "Database host")
	DBUser := flag.String("db-user", "", "Database user")
	DBPassword := flag.String("db-password", "", "Database pwd")
	DBScheme := flag.String("db-scheme", "", "Database scheme")
	flag.Parse()

	srvConfig.GRPC.Port = *GRPCPort
	srvConfig.Database.Host = *DBHost
	srvConfig.Database.User = *DBUser
	srvConfig.Database.Password = *DBPassword
	srvConfig.Database.Scheme = *DBScheme

	srv := v1.NewServer(srvConfig)
	if srv == nil {
		return fmt.Errorf("create server failed")
	}
	return srv.Run()
}

func RunServerV2() error {
	serviceConfig := &configs.ServiceConfig{}
	if err := configs.LoadConfig("", serviceConfig); err != nil {
		log.Fatalf("[Main] Load config failed: %v", err)
	}
	// if err := viper.Unmarshal(serviceConfig); err != nil {
	// 	log.Fatalf("[Main] Unmarshal config failed: %v", err)
	// }

	ctx := context.Background()

	// load config from cli
	// var config Config
	// flag.StringVar(&config.GRPCPort, "grpc-port", "", "gRPC port to bind")
	// flag.StringVar(&config.JWTSecretKey, "jwt-secret", "luloveyen", "jwt secret key")
	// flag.IntVar(&config.JWTDuration, "jwt-duration", 60, "jwt token duration (seconds)")
	// flag.Parse()

	v2API := v2.NewServiceAImpl()

	// create implementation extra service
	jwtManager := utils.NewJWTManager(serviceConfig.JWT)

	v2APIExtra := v2.NewServiceExtraImpl(jwtManager)

	// simple server interceptor
	// serverInterceptor := interceptor.SimpleServerInterceptor{}

	// auth server interceptor
	fmt.Printf("%+v\n", serviceConfig.AccessibleRoles)
	serverInterceptor := interceptor.NewAuthServerInterceptor(jwtManager, serviceConfig.AccessibleRoles)

	// auth with credentials interceptor
	// serverInterceptor := interceptor.NewCredentialsServerInterceptor(serviceConfig.Authentication)

	return grpc.RunServerV2(ctx, serverInterceptor, v2API, v2APIExtra, strconv.Itoa(serviceConfig.GRPC.Port))
}
