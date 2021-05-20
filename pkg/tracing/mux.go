package tracing

import (
	"context"
	"net/http"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/metadata"
)

type TraceServerMux struct {
	mux    *http.ServeMux
	tracer opentracing.Tracer
}

func NewTracerServerMux() *TraceServerMux {
	return &TraceServerMux{
		mux:    http.NewServeMux(),
		tracer: GlobalTracer(),
	}
}

func (t *TraceServerMux) SetTracer(tracer opentracing.Tracer) {
	t.tracer = tracer
}

func (t *TraceServerMux) Middleware(handler http.Handler) http.Handler {
	middleware := nethttp.Middleware(
		t.tracer,
		handler,
		nethttp.OperationNameFunc(func(r *http.Request) string {
			return "HTTP " + r.Method + r.URL.String()
		}),
		// nethttp.MWSpanObserver(func(sp opentracing.Span, r *http.Request) {
		// 	sp.SetTag("http.uri", r.URL.EscapedPath())
		// }),
	)
	return middleware
}

func (t *TraceServerMux) Handle(pattern string, handler http.Handler) {
	middleware := t.Middleware(handler)
	t.mux.Handle(pattern, middleware)
}

// Handler: implementation ServeHTTP
func (t *TraceServerMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.mux.ServeHTTP(w, r)
}

// inject spanCtx metadata into runtime server mux
func WithMetadata(ctx context.Context, r *http.Request) metadata.MD {
	span := opentracing.SpanFromContext(ctx)
	spanCtx := span.Context()
	carrier := make(map[string]string)
	if err := span.Tracer().Inject(
		spanCtx,
		opentracing.TextMap,
		opentracing.TextMapCarrier(carrier),
	); err != nil {
		return nil
	}
	return metadata.New(carrier)
}
