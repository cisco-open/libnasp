package tracing

import (
	"log"
	"net"
	"strconv"

	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	httpreporter "github.com/openzipkin/zipkin-go/reporter/http"

	"github.com/cisco-open/nasp/pkg/environment"
	"github.com/cisco-open/nasp/pkg/proxywasm/api"
)

// SetupZipkinTracing initializes and returns a zipkin tracer to be used for tracing middlewares and clients
func SetupZipkinTracing(istioEnvironment *environment.IstioEnvironment) (*zipkin.Tracer, error) {
	localIP := ""
	if len(istioEnvironment.InstanceIPs) > 0 {
		localIP = istioEnvironment.InstanceIPs[0]
	}

	localEndpoint, err := zipkin.NewEndpoint("nasp", localIP)
	if err != nil {
		return nil, err
	}

	reporter := httpreporter.NewReporter(istioEnvironment.ZipkinAddress)

	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(localEndpoint), zipkin.WithSharedSpans(false))
	if err != nil {
		log.Fatalf("unable to create tracer: %+v\n", err)
	}

	return tracer, nil
}

// ExtractRemoteEndpoint function extracts the remote endpoint information (service name, IPv4 address, and port number)
// from a given HTTPRequest
func ExtractRemoteEndpoint(r api.HTTPRequest) (*model.Endpoint, error) {
	// Extract remote IP and port from RemoteAddr
	remoteIP, remotePort, err := net.SplitHostPort(r.RemoteAddr())
	if err != nil {
		return nil, err
	}

	port, err := strconv.ParseUint(remotePort, 10, 16)
	if err != nil {
		return nil, err
	}

	// Extract service name from request URL
	url := r.URL().Host

	remoteEndpoint := &model.Endpoint{
		ServiceName: url,
		IPv4:        net.ParseIP(remoteIP),
		Port:        uint16(port),
	}

	return remoteEndpoint, nil
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

// InjectOption provides functional option handler type.
type InjectOption func(opts *InjectOptions)

// InjectOptions provides the available functional options.
type InjectOptions struct {
	shouldInjectSingleHeader bool
	shouldInjectMultiHeader  bool
}

// InjectHTTPHeaders will inject a span.Context into an HTTP Request
func InjectHTTPHeaders(r api.HTTPRequest, opts ...InjectOption) propagation.Injector {
	options := InjectOptions{shouldInjectMultiHeader: true}
	for _, opt := range opts {
		opt(&options)
	}

	return func(sc model.SpanContext) error {
		if (model.SpanContext{}) == sc {
			return b3.ErrEmptyContext
		}

		//nolint:nestif
		if options.shouldInjectMultiHeader {
			if sc.Debug {
				r.Header().Set(b3.Flags, "1")
			} else if sc.Sampled != nil {
				// Debug is encoded as X-B3-Flags: 1. Since Debug implies Sampled,
				// so don't also send "X-B3-Sampled: 1".
				if *sc.Sampled {
					r.Header().Set(b3.Sampled, "1")
				} else {
					r.Header().Set(b3.Sampled, "0")
				}
			}

			if !sc.TraceID.Empty() && sc.ID > 0 {
				r.Header().Set(b3.TraceID, sc.TraceID.String())
				r.Header().Set(b3.SpanID, sc.ID.String())
				if sc.ParentID != nil {
					r.Header().Set(b3.ParentSpanID, sc.ParentID.String())
				}
			}
		}

		if options.shouldInjectSingleHeader {
			r.Header().Set(b3.Context, b3.BuildSingleHeader(sc))
		}

		return nil
	}
}

func ExtractGRPCHeaders(req api.HTTPRequest) propagation.Extractor {
	return func() (*model.SpanContext, error) {
		var (
			traceIDHeader, _      = req.Header().Get(b3.TraceID)
			spanIDHeader, _       = req.Header().Get(b3.SpanID)
			parentSpanIDHeader, _ = req.Header().Get(b3.ParentSpanID)
			sampledHeader, _      = req.Header().Get(b3.Sampled)
			flagsHeader, _        = req.Header().Get(b3.Flags)
		)

		return b3.ParseHeaders(
			traceIDHeader, spanIDHeader, parentSpanIDHeader, sampledHeader,
			flagsHeader,
		)
	}
}

func InjectGRPCHeaders(req api.HTTPRequest) propagation.Injector {
	return func(sc model.SpanContext) error {
		if (model.SpanContext{}) == sc {
			return b3.ErrEmptyContext
		}

		if sc.Debug {
			req.Header().Set(b3.Flags, "1")
		} else if sc.Sampled != nil {
			if *sc.Sampled {
				req.Header().Set(b3.Sampled, "1")
			} else {
				req.Header().Set(b3.Sampled, "0")
			}
		}

		if !sc.TraceID.Empty() && sc.ID > 0 {
			// set identifiers
			req.Header().Set(b3.TraceID, sc.TraceID.String())
			req.Header().Set(b3.SpanID, sc.ID.String())
			if sc.ParentID != nil {
				req.Header().Set(b3.ParentSpanID, sc.ParentID.String())
			}
		}

		return nil
	}
}
