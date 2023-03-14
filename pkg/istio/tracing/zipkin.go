package tracing

import (
	"log"
	"os"

	"github.com/openzipkin/zipkin-go"
	httpreporter "github.com/openzipkin/zipkin-go/reporter/http"
)

// SetupZipkinTracing sets up and returns a zipkin tracer to be used for tracing middlewares and clients
func SetupZipkinTracing() *zipkin.Tracer {
	// TODO(bertab) Get zipkin endpoint from xDS
	endpoint, err := zipkin.NewEndpoint("zipkinService", os.Getenv("ZIPKIN_URL"))
	if err != nil {
		log.Fatalf("unable to create local endpoint: %+v\n", err)
	}

	reporter := httpreporter.NewReporter(string(endpoint.IPv4))
	defer func() {
		_ = reporter.Close()
	}()

	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		log.Fatalf("unable to create tracer: %+v\n", err)
	}

	return tracer
}
