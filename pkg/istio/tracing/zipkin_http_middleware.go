package tracing

import (
	"strconv"

	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation"
	"github.com/openzipkin/zipkin-go/propagation/b3"

	"github.com/cisco-open/nasp/pkg/proxywasm/api"
	lhttp "github.com/cisco-open/nasp/pkg/proxywasm/http"
)

type zipkinHTTPTracingMiddleware struct {
	tracer *zipkin.Tracer

	span zipkin.Span
}

func NewZipkinHTTPTracingMiddleware(tracer *zipkin.Tracer) lhttp.HandleMiddleware {
	h := &zipkinHTTPTracingMiddleware{
		tracer: tracer,
	}
	return h
}

func (z *zipkinHTTPTracingMiddleware) BeforeRequest(req api.HTTPRequest, stream api.Stream) {
	// extract B3 Headers from upstream
	spanContext := z.tracer.Extract(ExtractHTTPHeaders(req))

	spanName := req.Method()

	z.span = z.tracer.StartSpan(
		spanName,
		zipkin.Kind(model.Server),
		zipkin.Parent(spanContext),
	)

	// add our span to context
	ctx := zipkin.NewContext(req.Context(), z.span)

	if zipkin.IsNoop(z.span) {
		// while the span is not being recorded, we still want to propagate the context.
		req = req.WithContext(ctx)
		return
	}

	remoteEndpoint, _ := zipkin.NewEndpoint("", req.Connection().RemoteAddr().String())
	z.span.SetRemoteEndpoint(remoteEndpoint)

	// tag typical HTTP request items
	zipkin.TagHTTPMethod.Set(z.span, req.Method())
	zipkin.TagHTTPPath.Set(z.span, req.URL().Path)
	if req.ContentLength() > 0 {
		zipkin.TagHTTPRequestSize.Set(z.span, strconv.FormatInt(req.ContentLength(), 10))
	}

	// continue using our updated context.
	req = req.WithContext(ctx)
}

func (z *zipkinHTTPTracingMiddleware) AfterRequest(req api.HTTPRequest, stream api.Stream) {}

func (z *zipkinHTTPTracingMiddleware) BeforeResponse(resp api.HTTPResponse, stream api.Stream) {}

func (z *zipkinHTTPTracingMiddleware) AfterResponse(resp api.HTTPResponse, stream api.Stream) {
	code := resp.StatusCode()
	sCode := strconv.Itoa(code)
	if code < 200 || code > 299 {
		zipkin.TagHTTPStatusCode.Set(z.span, sCode)
		if code > 399 {
			zipkin.TagError.Set(z.span, sCode)
		}
	}
	if resp.ContentLength() > 0 {
		zipkin.TagHTTPResponseSize.Set(z.span, strconv.FormatInt(resp.ContentLength(), 10))
	}
	z.span.Finish()
}

// ExtractHTTPHeaders will extract a span.Context from the HTTP Request if found in
// B3 header format.
func ExtractHTTPHeaders(r api.HTTPRequest) propagation.Extractor {
	return func() (*model.SpanContext, error) {
		var (
			traceIDHeader, _      = r.Header().Get(b3.TraceID)
			spanIDHeader, _       = r.Header().Get(b3.SpanID)
			parentSpanIDHeader, _ = r.Header().Get(b3.ParentSpanID)
			sampledHeader, _      = r.Header().Get(b3.Sampled)
			flagsHeader, _        = r.Header().Get(b3.Flags)
			singleHeader, _       = r.Header().Get(b3.Context)
		)

		var (
			sc   *model.SpanContext
			sErr error
			mErr error
		)
		if singleHeader != "" {
			sc, sErr = b3.ParseSingleHeader(singleHeader)
			if sErr == nil {
				return sc, nil
			}
		}

		sc, mErr = b3.ParseHeaders(
			traceIDHeader, spanIDHeader, parentSpanIDHeader,
			sampledHeader, flagsHeader,
		)

		if mErr != nil && sErr != nil {
			return nil, sErr
		}

		return sc, mErr
	}
}
