package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/viper"
	"github.com/unrolled/secure"
	"google.golang.org/grpc"

	api_v2 "grpc-rest-microservice/pkg/api/v2"
	"grpc-rest-microservice/pkg/configs"
)

var (
	gRPCHost  string
	proxyPort string
	demo      = flag.String("demo-grpc", "lol", "demo grpc argument")
)

func InitRouter(handler http.Handler) *gin.Engine {
	if os.Getenv("GOENV") != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}

	secureMiddleware := secure.New(secure.Options{
		AllowedHosts: []string{
			"*",
		},
		AllowedHostsAreRegex:  true,
		HostsProxyHeaders:     []string{"X-Forwarded-Host"},
		SSLRedirect:           false,
		SSLHost:               "",
		SSLProxyHeaders:       map[string]string{},
		STSSeconds:            31536000,
		STSIncludeSubdomains:  true,
		STSPreload:            true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'; img-src 'self' data:; media-src 'self' data:; font-src 'self' data:",
		PublicKey:             `pin-sha256="base64+primary=="; pin-sha256="base64+backup=="; max-age=5184000; includeSubdomains; report-uri="http://localhost:8082"`,
	})

	secureFunc := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			err := secureMiddleware.Process(c.Writer, c.Request)

			// If there was an error, do not continue.
			if err != nil {
				log.Println("err", err)
				c.Abort()
				return
			}

			// Avoid header rewrite if response is a redirection.
			if status := c.Writer.Status(); status > 300 && status < 399 {
				log.Println("status", status)
				c.Abort()
			}
		}
	}()

	r := gin.Default()
	r.Use(secureFunc)

	r.Any("/*aaa", gin.WrapH(handler))

	// r.GET("/", func(c *gin.Context) {
	// 	c.String(200, "Have nice day")
	// })

	return r
}

func main() {

	serviceConfig := &configs.ServiceConfig{}
	if err := configs.LoadConfig(); err != nil {
		log.Fatalf("[Main] Load config failed: %v", err)
	}
	if err := viper.Unmarshal(serviceConfig); err != nil {
		log.Fatalf("[Main] Unmarshal config failed: %v", err)
	}

	flag.StringVar(&gRPCHost, "grpc-host", "", "gRPC host to bind")
	flag.StringVar(&proxyPort, "proxy-port", "", "proxy port to bind")
	flag.Parse()

	// fmt.Println(gRPCHost, proxyPort, *demo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()

	gRPCHost = fmt.Sprintf("%s:%d", serviceConfig.GRPC.Host, serviceConfig.GRPC.Port)
	err := api_v2.RegisterServiceAHandlerFromEndpoint(
		ctx, mux, gRPCHost,
		[]grpc.DialOption{grpc.WithInsecure()},
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connect gRPC at:", gRPCHost)

	// err = http.ListenAndServe(":"+proxyPort, mux)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	proxyPort = fmt.Sprintf("%d", serverConfig.Proxy.Port)
	r := InitRouter(mux)
	log.Println("Proxy gw running at:", proxyPort)
	if err := r.Run(":" + proxyPort); err != nil {
		log.Fatalf("Running grpc gateway error: %+v", err)
	}
}
