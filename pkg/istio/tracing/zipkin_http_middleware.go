package tracing

import (
	"context"
	"strconv"

	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"

	"github.com/cisco-open/nasp/pkg/proxywasm/api"
	lhttp "github.com/cisco-open/nasp/pkg/proxywasm/http"
)

type zipkinHTTPTracingMiddleware struct {
	tracer *zipkin.Tracer
}

var _ lhttp.HandleMiddleware = &zipkinHTTPTracingMiddleware{}

func NewZipkinHTTPTracingMiddleware(tracer *zipkin.Tracer) lhttp.HandleMiddleware {
	h := &zipkinHTTPTracingMiddleware{
		tracer: tracer,
	}
	return h
}

func (z *zipkinHTTPTracingMiddleware) BeforeRequest(req api.HTTPRequest, stream api.Stream) (api.HTTPRequest, api.Stream) {
	if z.tracer == nil {
		return req, stream
	}

	// extract B3 Headers from upstream
	spanContext := z.tracer.Extract(ExtractHTTPHeaders(req))

	spanName := req.URL().Scheme + "/" + req.Method()

	span := z.tracer.StartSpan(
		spanName,
		zipkin.Kind(model.Server),
		zipkin.Parent(spanContext),
	)

	remoteEndpoint, _ := ExtractRemoteEndpoint(req)
	span.SetRemoteEndpoint(remoteEndpoint)

	// tag typical HTTP request items
	zipkin.TagHTTPMethod.Set(span, req.Method())
	zipkin.TagHTTPPath.Set(span, req.URL().Path)
	zipkin.TagHTTPUrl.Set(span, req.URL().String())

	if req.ContentLength() > 0 {
		zipkin.TagHTTPRequestSize.Set(span, strconv.FormatInt(req.ContentLength(), 10))
	}

	_ = InjectHTTPHeaders(req)(span.Context())

	return req.WithContext(zipkin.NewContext(req.Context(), span)), stream
}

func (z *zipkinHTTPTracingMiddleware) AfterRequest(req api.HTTPRequest, stream api.Stream) (api.HTTPRequest, api.Stream) {
	return req, stream
}

func (z *zipkinHTTPTracingMiddleware) BeforeResponse(ctx context.Context, resp api.HTTPResponse, stream api.Stream) (api.HTTPResponse, api.Stream) {
	return resp, stream
}

func (z *zipkinHTTPTracingMiddleware) AfterResponse(ctx context.Context, resp api.HTTPResponse, stream api.Stream) (api.HTTPResponse, api.Stream) {
	if z.tracer == nil {
		return resp, stream
	}

	if span := zipkin.SpanFromContext(ctx); span != nil {
		code := resp.StatusCode()
		sCode := strconv.Itoa(code)
		if code < 200 || code > 299 {
			zipkin.TagHTTPStatusCode.Set(span, sCode)
			if code > 399 {
				zipkin.TagError.Set(span, sCode)
			}
		}
		if resp.ContentLength() > 0 {
			zipkin.TagHTTPResponseSize.Set(span, strconv.FormatInt(resp.ContentLength(), 10))
		}

		span.Finish()
	}

	return resp, stream
}
