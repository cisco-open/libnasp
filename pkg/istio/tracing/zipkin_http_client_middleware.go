package tracing

import (
	"context"
	"strconv"

	"github.com/cisco-open/nasp/pkg/environment"

	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"

	"github.com/cisco-open/nasp/pkg/proxywasm/api"
	lhttp "github.com/cisco-open/nasp/pkg/proxywasm/http"
)

type zipkinHTTPClientTracingMiddleware struct {
	tracer      *zipkin.Tracer
	environment *environment.IstioEnvironment
}

var _ lhttp.HandleMiddleware = &zipkinHTTPClientTracingMiddleware{}

func NewZipkinHTTPClientTracingMiddleware(tracer *zipkin.Tracer, meshConfig *environment.IstioEnvironment) lhttp.HandleMiddleware {
	h := &zipkinHTTPClientTracingMiddleware{
		tracer:      tracer,
		environment: meshConfig,
	}
	return h
}

func (z zipkinHTTPClientTracingMiddleware) BeforeRequest(req api.HTTPRequest, stream api.Stream) (api.HTTPRequest, api.Stream) {
	return req, stream
}

func (z zipkinHTTPClientTracingMiddleware) AfterRequest(req api.HTTPRequest, stream api.Stream) (api.HTTPRequest, api.Stream) {
	if z.tracer == nil {
		return req, stream
	}

	var parentContext model.SpanContext

	if span := zipkin.SpanFromContext(req.Context()); span != nil {
		parentContext = span.Context()
	} else {
		parentContext = z.tracer.Extract(ExtractHTTPHeaders(req))
	}

	spanName := req.URL().String()
	remoteEndpoint, _ := ExtractRemoteEndpoint(req)

	span := z.tracer.StartSpan(
		spanName,
		zipkin.Kind(model.Client),
		zipkin.Parent(parentContext),
		zipkin.RemoteEndpoint(remoteEndpoint))

	zipkin.TagHTTPMethod.Set(span, req.Method())
	zipkin.TagHTTPPath.Set(span, req.URL().Path)
	zipkin.TagHTTPUrl.Set(span, req.URL().String())
	AddCommonZipkinTags(req, span, z.environment)

	_ = InjectHTTPHeaders(req)(span.Context())

	return req.WithContext(zipkin.NewContext(req.Context(), span)), stream
}

func (z zipkinHTTPClientTracingMiddleware) BeforeResponse(ctx context.Context, resp api.HTTPResponse, stream api.Stream) (api.HTTPResponse, api.Stream) {
	return resp, stream
}

func (z zipkinHTTPClientTracingMiddleware) AfterResponse(ctx context.Context, resp api.HTTPResponse, stream api.Stream) (api.HTTPResponse, api.Stream) {
	if z.tracer == nil {
		return resp, stream
	}

	if span := zipkin.SpanFromContext(ctx); span != nil {
		if resp.ContentLength() > 0 {
			zipkin.TagHTTPResponseSize.Set(span, strconv.FormatInt(resp.ContentLength(), 10))
		}
		if resp.StatusCode() < 200 || resp.StatusCode() > 299 {
			statusCode := strconv.FormatInt(int64(resp.StatusCode()), 10)
			zipkin.TagHTTPStatusCode.Set(span, statusCode)
			if resp.StatusCode() > 399 {
				zipkin.TagError.Set(span, statusCode)
			}
		}
		span.Finish()
	}

	return resp, stream
}
