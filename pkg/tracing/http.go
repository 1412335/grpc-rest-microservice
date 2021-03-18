package tracing

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
)

type HTTPClient struct {
	Client *http.Client
	Tracer opentracing.Tracer
}

// Do executes an HTTP request and returns the response body.
// Any errors or non-200 status code result in an error.
func (h *HTTPClient) Do(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req, ht := nethttp.TraceRequest(h.Tracer, req, nethttp.OperationName("HTTP GET: "+url))
	defer ht.Finish()

	// pass context to child service
	// span.kind=client
	// ext.SpanKindRPCClient.Set(span)
	// ext.HTTPUrl.Set(span, url)
	// ext.HTTPMethod.Set(span, "GET")
	// span.Tracer().Inject(
	// 	span.Context(),
	// 	opentracing.HTTPHeaders,
	// 	opentracing.HTTPHeadersCarrier(req.Header),
	// )

	// do request to formatter service
	resp, err := h.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("StatusCode: %d, Body: %s", resp.StatusCode, body)
	}

	return body, nil
}
