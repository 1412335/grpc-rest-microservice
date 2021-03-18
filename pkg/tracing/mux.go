package tracing

import (
	"net/http"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
)

type TraceServerMux struct {
	mux    *http.ServeMux
	tracer opentracing.Tracer
}

func NewTracerServerMux(tracer opentracing.Tracer) *TraceServerMux {
	return &TraceServerMux{
		mux:    http.NewServeMux(),
		tracer: tracer,
	}
}

func (t *TraceServerMux) Handle(pattern string, handler http.Handler) {
	middleware := nethttp.Middleware(
		t.tracer,
		handler,
		nethttp.OperationNameFunc(func(r *http.Request) string {
			return "HTTP " + r.Method + pattern
		}),
		// nethttp.MWSpanObserver(func(sp opentracing.Span, r *http.Request) {
		// 	sp.SetTag("http.uri", r.URL.EscapedPath())
		// }),
	)
	t.mux.Handle(pattern, middleware)
}

// Handler: implementation ServeHTTP
func (t *TraceServerMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.mux.ServeHTTP(w, r)
}
