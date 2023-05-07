package tracing

import (
	"context"
	"strconv"

	"google.golang.org/grpc/status"

	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"

	"github.com/cisco-open/nasp/pkg/proxywasm/api"
	lhttp "github.com/cisco-open/nasp/pkg/proxywasm/http"
)

type zipkinGRPCTracingMiddleware struct {
	tracer *zipkin.Tracer
}

var _ lhttp.HandleMiddleware = &zipkinGRPCTracingMiddleware{}

func NewZipkinGRPCTracingMiddleware(tracer *zipkin.Tracer) lhttp.HandleMiddleware {
	h := &zipkinGRPCTracingMiddleware{
		tracer: tracer,
	}
	return h
}

func (z zipkinGRPCTracingMiddleware) BeforeRequest(req api.HTTPRequest, stream api.Stream) (api.HTTPRequest, api.Stream) {
	tags := make(map[string]string)

	tags["grpc.method"] = req.Method()

	spanName := req.URL().String()
	remoteEndpoint, _ := ExtractRemoteEndpoint(req)

	spanContext := z.tracer.Extract(ExtractGRPCHeaders(req))

	span := z.tracer.StartSpan(
		spanName,
		zipkin.Kind(model.Server),
		zipkin.Tags(tags),
		zipkin.Parent(spanContext),
		zipkin.RemoteEndpoint(remoteEndpoint),
	)
	span.SetRemoteEndpoint(remoteEndpoint)

	if req.ContentLength() > 0 {
		zipkin.TagHTTPRequestSize.Set(span, strconv.FormatInt(req.ContentLength(), 10))
	}

	_ = InjectGRPCHeaders(req)(span.Context())
	return req.WithContext(zipkin.NewContext(req.Context(), span)), stream
}

func (z zipkinGRPCTracingMiddleware) AfterRequest(req api.HTTPRequest, stream api.Stream) (api.HTTPRequest, api.Stream) {
	return req, stream
}

func (z zipkinGRPCTracingMiddleware) BeforeResponse(ctx context.Context, resp api.HTTPResponse, stream api.Stream) (api.HTTPResponse, api.Stream) {
	return resp, stream
}

func (z zipkinGRPCTracingMiddleware) AfterResponse(ctx context.Context, resp api.HTTPResponse, stream api.Stream) (api.HTTPResponse, api.Stream) {
	if span := zipkin.SpanFromContext(ctx); span != nil {
		if resp.Error() != nil {
			if status, ok := status.FromError(resp.Error()); ok {
				zipkin.TagGRPCStatusCode.Set(span, strconv.Itoa(resp.StatusCode()))
				zipkin.TagError.Set(span, status.Message())
			} else {
				zipkin.TagError.Set(span, resp.Error().Error())
			}
		}

		span.Finish()
	}

	return resp, stream
}
