package tracing

import (
	"context"

	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"

	"github.com/cisco-open/nasp/pkg/proxywasm/api"
	lhttp "github.com/cisco-open/nasp/pkg/proxywasm/http"
)

type zipkinGRPCClientTracingMiddleware struct {
	tracer *zipkin.Tracer
}

var _ lhttp.HandleMiddleware = &zipkinGRPCClientTracingMiddleware{}

func NewZipkinGRPCClientTracingMiddleware(tracer *zipkin.Tracer) lhttp.HandleMiddleware {
	h := &zipkinGRPCClientTracingMiddleware{
		tracer: tracer,
	}
	return h
}

func (z zipkinGRPCClientTracingMiddleware) BeforeRequest(req api.HTTPRequest, stream api.Stream) (api.HTTPRequest, api.Stream) {
	return req, stream
}

func (z zipkinGRPCClientTracingMiddleware) AfterRequest(req api.HTTPRequest, stream api.Stream) (api.HTTPRequest, api.Stream) {
	if z.tracer == nil {
		return req, stream
	}

	var parentContext model.SpanContext

	spanName := req.URL().String()

	if parent := zipkin.SpanFromContext(req.Context()); parent != nil {
		parentContext = parent.Context()
	}

	remoteEndpoint, _ := ExtractRemoteEndpoint(req)

	span := z.tracer.StartSpan(
		spanName,
		zipkin.Kind(model.Client),
		zipkin.Parent(parentContext),
		zipkin.RemoteEndpoint(remoteEndpoint),
	)

	_ = InjectGRPCHeaders(req)(span.Context())

	return req.WithContext(zipkin.NewContext(req.Context(), span)), stream
}

func (z zipkinGRPCClientTracingMiddleware) BeforeResponse(ctx context.Context, resp api.HTTPResponse, stream api.Stream) (api.HTTPResponse, api.Stream) {
	return resp, stream
}

func (z zipkinGRPCClientTracingMiddleware) AfterResponse(ctx context.Context, resp api.HTTPResponse, stream api.Stream) (api.HTTPResponse, api.Stream) {
	if z.tracer == nil {
		return resp, stream
	}

	if span := zipkin.SpanFromContext(ctx); span != nil {
		if resp.Error() != nil {
			zipkin.TagError.Set(span, resp.Error().Error())
		}

		span.Finish()
	}

	return resp, stream
}
