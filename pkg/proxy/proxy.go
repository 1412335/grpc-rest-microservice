package proxy

import (
	"context"
	"log"
	"mime"
	"net"
	"net/http"
	"net/textproto"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/rakyll/statik/fs"
	"github.com/unrolled/secure"
	"google.golang.org/grpc"

	api_v2 "github.com/1412335/grpc-rest-microservice/pkg/api/v2/grpc-gateway/gen"
	"github.com/1412335/grpc-rest-microservice/pkg/configs"

	// Static files
	_ "github.com/1412335/grpc-rest-microservice/pkg/api/v2/grpc-gateway/statik"
)

type handler struct {
	config *configs.ServiceConfig
}

func NewHandler(config *configs.ServiceConfig) *handler {
	return &handler{
		config: config,
	}
}

// isPermanentHTTPHeader checks whether hdr belongs to the list of
// permanent request headers maintained by IANA.
// http://www.iana.org/assignments/message-headers/message-headers.xml
// From https://github.com/grpc-ecosystem/grpc-gateway/blob/7a2a43655ccd9a488d423ea41a3fc723af103eda/runtime/context.go#L157
func (h *handler) isPermanentHTTPHeader(hdr string) bool {
	switch hdr {
	case
		"Accept",
		"Accept-Charset",
		"Accept-Language",
		"Accept-Ranges",
		"Authorization",
		"Cache-Control",
		"Content-Type",
		"Cookie",
		"Date",
		"Expect",
		"From",
		"Host",
		"If-Match",
		"If-Modified-Since",
		"If-None-Match",
		"If-Schedule-Tag-Match",
		"If-Unmodified-Since",
		"Max-Forwards",
		"Origin",
		"Pragma",
		"Referer",
		"User-Agent",
		"Via",
		"Warning":
		return true
	}
	return false
}

// isReserved returns whether the key is reserved by gRPC.
func (h *handler) isReserved(key string) bool {
	return strings.HasPrefix(key, "Grpc-")
}

// incomingHeaderMatcher converts an HTTP header name on http.Request to
// grpc metadata. Permanent headers (i.e. User-Agent) are prepended with
// "grpc-gateway". Headers that start with start with "Grpc-" (reserved
// by grpc) are prepended with "X-". Other headers are forwarded as is.
func (h *handler) incomingHeaderMatcher(key string) (string, bool) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	if h.isPermanentHTTPHeader(key) {
		return runtime.MetadataPrefix + key, true
	}
	if h.isReserved(key) {
		return "X-" + key, true
	}

	// The Istio service mesh dislikes when you pass the Content-Length header
	if key == "Content-Length" {
		return "", false
	}

	return key, true
}

// outgoingHeaderMatcher transforms outgoing metadata into HTTP headers.
// We return any response metadata as is.
func (h *handler) outgoingHeaderMatcher(metadata string) (string, bool) {
	return metadata, true
}

// init gin router
func (h *handler) initRouter(handler http.Handler) *gin.Engine {
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
		ContentSecurityPolicy: "default-src 'self' 'unsafe-inline'; img-src 'self' data:; media-src 'self' data:; font-src 'self' data:",
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

	if err := serveOpenAPI(r); err != nil {
		log.Println("serveOpenAPI failed:", err)
	}

	// r.GET("/", func(c *gin.Context) {
	// 	c.String(http.StatusOK, "Have nice day")
	// })

	api := r.Group("/v2")
	api.Any("/*any", gin.WrapH(handler))

	return r
}

// serveOpenAPI serves an OpenAPI UI on /openapi-ui/
// Adapted from https://github.com/philips/grpc-gateway-example/blob/a269bcb5931ca92be0ceae6130ac27ae89582ecc/cmd/serve.go#L63
func serveOpenAPI(r *gin.Engine) error {
	mime.AddExtensionType(".svg", "image/svg+xml")
	statikFS, err := fs.New()
	if err != nil {
		return err
	}
	// Expose files in static on <host>/openapi-ui
	// fileServer := http.FileServer(statikFS)
	prefix := "/openapi-ui/"
	r.StaticFS(prefix, statikFS)
	// r.GET(prefix, gin.WrapH(http.StripPrefix(prefix, fileServer))) => not working
	// r.Static("/openui", "pkg/api/v2/grpc-gateway/third_party/OpenAPI")
	return nil
}

// run grpc-gateway
func (h *handler) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(h.incomingHeaderMatcher),
		runtime.WithOutgoingHeaderMatcher(h.outgoingHeaderMatcher),
	)

	// gRPCHost := net.JoinHostPort(h.config.GRPC.Host, strconv.Itoa(h.config.GRPC.Port))
	gRPCHost := net.JoinHostPort("localhost", strconv.Itoa(h.config.GRPC.Port))

	// register handler
	err := api_v2.RegisterServiceAHandlerFromEndpoint(
		ctx,
		mux,
		gRPCHost,
		[]grpc.DialOption{grpc.WithInsecure()},
	)
	if err != nil {
		return err
	}

	// proxy address
	addr := ":" + strconv.Itoa(h.config.Proxy.Port)
	// router
	router := h.initRouter(mux)
	// http server
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case sig := <-signals:
			log.Println("Proxy gateway signal received:", sig)
			shutdown, can := context.WithTimeout(ctx, 10*time.Second)
			srv.Shutdown(shutdown)
			defer can()
		}
	}()

	log.Println("Proxy gateway running at:", addr)
	return srv.ListenAndServe()
}
