package v2

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/unrolled/secure"
	"google.golang.org/grpc"

	api_v2 "github.com/1412335/grpc-rest-microservice/pkg/api/v2/gen/grpc-gateway/gen"
	"github.com/1412335/grpc-rest-microservice/pkg/configs"
)

type client struct {
	config *configs.ServiceConfig
}

func NewClient(config *configs.ServiceConfig) *client {
	return &client{
		config: config,
	}
}

func (c *client) initRouter(handler http.Handler) *gin.Engine {
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

	// r.GET("/", func(c *gin.Context) {
	// 	c.String(http.StatusOK, "Have nice day")
	// })

	r.GET("/", gin.WrapH(handler))

	return r
}

func (c *client) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()

	gRPCHost := net.JoinHostPort(c.config.GRPC.Host, strconv.Itoa(c.config.GRPC.Port))
	err := api_v2.RegisterServiceAHandlerFromEndpoint(
		ctx,
		mux,
		gRPCHost,
		[]grpc.DialOption{grpc.WithInsecure()},
	)
	if err != nil {
		return err
	}

	proxyPort := strconv.Itoa(c.config.Proxy.Port)
	r := c.initRouter(mux)
	log.Println("Proxy gateway running at:", proxyPort)
	return r.Run(":" + proxyPort)
}
