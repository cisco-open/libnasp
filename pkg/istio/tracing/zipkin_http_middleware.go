package tracing

import (
	"context"
	"strconv"

	"google.golang.org/grpc/status"

	"github.com/cisco-open/nasp/pkg/environment"

	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"

	"github.com/cisco-open/nasp/pkg/proxywasm/api"
	lhttp "github.com/cisco-open/nasp/pkg/proxywasm/http"
)

type zipkinHTTPTracingMiddleware struct {
	tracer      *zipkin.Tracer
	environment *environment.IstioEnvironment
}

var _ lhttp.HandleMiddleware = &zipkinHTTPTracingMiddleware{}

func NewZipkinHTTPTracingMiddleware(tracer *zipkin.Tracer, meshConfig *environment.IstioEnvironment) lhttp.HandleMiddleware {
	h := &zipkinHTTPTracingMiddleware{
		tracer:      tracer,
		environment: meshConfig,
	}
	return h
}

func (z *zipkinHTTPTracingMiddleware) BeforeRequest(req api.HTTPRequest, stream api.Stream) (api.HTTPRequest, api.Stream) {
	if z.tracer == nil {
		return req, stream
	}

	// handle GRPC request
	if value, ok := req.Header().Get("Content-Type"); ok && value == "application/grpc" {
		return z.GRPCBeforeRequest(req, stream)
	}

	// extract B3 Headers from upstream
	spanContext := z.tracer.Extract(ExtractHTTPHeaders(req))

	spanName := req.URL().String()

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
	AddCommonZipkinTags(req, span, z.environment)

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

	// handle GRPC case
	if value, ok := resp.Header().Get("Content-Type"); ok && value == "application/grpc" {
		return z.GRPCAfterResponse(ctx, resp, stream)
	}

	if span := zipkin.SpanFromContext(ctx); span != nil {
		code := resp.StatusCode()
		sCode := strconv.Itoa(code)
		if code <= 200 || code > 299 {
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

func (z *zipkinHTTPTracingMiddleware) GRPCBeforeRequest(req api.HTTPRequest, stream api.Stream) (api.HTTPRequest, api.Stream) {
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

func (z *zipkinHTTPTracingMiddleware) GRPCAfterResponse(ctx context.Context, resp api.HTTPResponse, stream api.Stream) (api.HTTPResponse, api.Stream) {
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

// AddCommonZipkinTags adds common tracing tags, like HTTP protocol, user agent, request ID, downstream cluster,
// and Istio environment specifics
func AddCommonZipkinTags(req api.HTTPRequest, span zipkin.Span, istioEnvironment *environment.IstioEnvironment) {
	zipkin.Tag("http.protocol").Set(span, req.HTTPProtocol())
	zipkin.Tag("component").Set(span, "nasp")
	if value, ok := req.Header().Get("User-Agent"); ok {
		zipkin.Tag("user_agent").Set(span, value)
	}
	if value, ok := req.Header().Get("x-request-id"); ok {
		zipkin.Tag("guid:x-request-id").Set(span, value)
	}

	if value, ok := req.Header().Get("x-envoy-downstream-service-cluster"); ok {
		zipkin.Tag("downstream_cluster").Set(span, value)
	}

	zipkin.Tag("istio.mesh_id").Set(span, istioEnvironment.MeshID)
	zipkin.Tag("istio.namespace").Set(span, istioEnvironment.PodNamespace)

	if value, ok := istioEnvironment.Labels["service.istio.io/canonical-name"]; ok {
		zipkin.Tag("istio.canonical_service").Set(span, value)
	}

	if value, ok := istioEnvironment.Labels["service.istio.io/canonical-revision"]; ok {
		zipkin.Tag("istio.canonical_revision").Set(span, value)
	}

	if req.ContentLength() > 0 {
		zipkin.TagHTTPRequestSize.Set(span, strconv.FormatInt(req.ContentLength(), 10))
	}
}
