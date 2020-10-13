package cmd

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strconv"

	"grpc-rest-microservice/pkg/configs"
	"grpc-rest-microservice/pkg/interceptor"
	"grpc-rest-microservice/pkg/protocol/grpc"
	v1 "grpc-rest-microservice/pkg/service/v1"
	v2 "grpc-rest-microservice/pkg/service/v2"
	"grpc-rest-microservice/pkg/utils"

	"github.com/spf13/viper"
)

func RunServer() error {

	serviceConfig := &configs.ServiceConfig{}
	if err := configs.LoadConfig(); err != nil {
		log.Fatalf("[Main] Load config failed: %v", err)
	}
	if err := viper.Unmarshal(serviceConfig); err != nil {
		log.Fatalf("[Main] Unmarshal config failed: %v", err)
	}

	GRPCPort := flag.Int("grpc-port", 9090, "gRPC port service")
	DBHost := flag.String("db-host", "", "Database host")
	DBUser := flag.String("db-user", "", "Database user")
	DBPassword := flag.String("db-password", "", "Database pwd")
	DBScheme := flag.String("db-scheme", "", "Database scheme")
	flag.Parse()

	serviceConfig.GRPC.Port = *GRPCPort
	serviceConfig.Database.Host = *DBHost
	serviceConfig.Database.User = *DBUser
	serviceConfig.Database.Password = *DBPassword
	serviceConfig.Database.Scheme = *DBScheme

	ctx := context.Background()

	param := "parseTime=true"
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s",
		serviceConfig.Database.User,
		serviceConfig.Database.Password,
		serviceConfig.Database.Host,
		serviceConfig.Database.Scheme,
		param,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect database: %v", err)
	}
	defer db.Close()

	v1API := v1.NewToDoServiceServer(db)

	return grpc.RunServer(ctx, v1API, strconv.Itoa(serviceConfig.GRPC.Port))
}

func RunServerV2() error {

	serviceConfig := &configs.ServiceConfig{}
	if err := configs.LoadConfig(); err != nil {
		log.Fatalf("[Main] Load config failed: %v", err)
	}
	if err := viper.Unmarshal(serviceConfig); err != nil {
		log.Fatalf("[Main] Unmarshal config failed: %v", err)
	}

	ctx := context.Background()

	// var config Config
	// flag.StringVar(&config.GRPCPort, "grpc-port", "", "gRPC port to bind")
	// flag.StringVar(&config.JWTSecretKey, "jwt-secret", "luloveyen", "jwt secret key")
	// flag.IntVar(&config.JWTDuration, "jwt-duration", 60, "jwt token duration (seconds)")
	// flag.Parse()

	v2API := v2.NewServiceAImpl()

	// create implementation extra service
	jwtManager := utils.NewJWTManager(serviceConfig.JWT)

	v2API_extra := v2.NewServiceExtraImpl(jwtManager)

	// simple server interceptor
	// serverInterceptor := interceptor.SimpleServerInterceptor{}

	// auth server interceptor
	fmt.Printf("%+v\n", serviceConfig.AccessibleRoles)
	serverInterceptor := interceptor.NewAuthServerInterceptor(jwtManager, serviceConfig.AccessibleRoles)

	// auth with credentials interceptor
	// serverInterceptor := interceptor.NewCredentialsServerInterceptor(serviceConfig.Authentication)

	return grpc.RunServerV2(ctx, serverInterceptor, v2API, v2API_extra, strconv.Itoa(serviceConfig.GRPC.Port))
}
