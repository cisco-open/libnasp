package tracing

import (
	"net"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	"github.com/openzipkin/zipkin-go/reporter"

	"github.com/banzaicloud/operator-tools/pkg/utils"
	"github.com/cisco-open/nasp/pkg/proxywasm/api"

	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	wrapper "github.com/cisco-open/nasp/pkg/proxywasm/http"
)

func TestZipkin(t *testing.T) {
	RegisterFailHandler(Fail)
	t.Parallel()
	RunSpecs(t, "Zipkin Suite")
}

func newMockRequestWithHeaders(headers map[string]string) api.HTTPRequest {
	request := &http.Request{Header: map[string][]string{}}
	mockRequest := wrapper.WrapHTTPRequest(request)

	for key, value := range headers {
		mockRequest.Header().Set(key, value)
	}

	return mockRequest
}

func createSpanContext(traceIDHex, spanIDHex string, sampled bool) (model.SpanContext, error) {
	traceID, err := model.TraceIDFromHex(traceIDHex)
	if err != nil {
		return model.SpanContext{}, err
	}
	spanID, err := strconv.ParseUint(spanIDHex, 16, 64)
	if err != nil {
		return model.SpanContext{}, err
	}
	return model.SpanContext{
		TraceID: traceID,
		ID:      model.ID(spanID),
		Sampled: utils.BoolPointer(sampled),
	}, nil
}

func startSpan() zipkin.Span {
	tracer, err := zipkin.NewTracer(reporter.NewNoopReporter())
	if err != nil {
		panic(err)
	}

	return tracer.StartSpan("new_test_span")
}

var _ = Describe("Propagation", func() {
	var span zipkin.Span

	BeforeEach(func() {
		span = startSpan()
	})
	AfterEach(func() {
		span.Finish()
	})
	Context("ExtractHTTPHeaders", func() {
		It("should extract a span context from the HTTP Request", func() {
			headers := map[string]string{
				b3.TraceID: span.Context().TraceID.String(),
				b3.SpanID:  span.Context().ID.String(),
				b3.Sampled: "1",
			}
			mockRequest := newMockRequestWithHeaders(headers)

			extractor := ExtractHTTPHeaders(mockRequest)
			sc, err := extractor()

			Expect(err).ToNot(HaveOccurred())
			Expect(sc.TraceID.String()).To(Equal(headers[b3.TraceID]))
			Expect(sc.ID.String()).To(Equal(headers[b3.SpanID]))
			Expect(*sc.Sampled).To(BeTrue())
		})

		It("should return a span context with debug flag when B3 debug header is present", func() {
			headers := map[string]string{
				b3.TraceID: span.Context().TraceID.String(),
				b3.SpanID:  span.Context().ID.String(),
				b3.Flags:   "1",
			}
			mockRequest := newMockRequestWithHeaders(headers)

			extractor := ExtractHTTPHeaders(mockRequest)
			sc, err := extractor()

			Expect(err).ToNot(HaveOccurred())
			Expect(sc.TraceID.String()).To(Equal(headers[b3.TraceID]))
			Expect(sc.ID.String()).To(Equal(headers[b3.SpanID]))
			Expect(sc.Debug).To(BeTrue())
		})
	})

	Context("InjectHTTPHeaders", func() {
		It("should inject a span context into a HTTP Request", func() {
			mockRequest := newMockRequestWithHeaders(nil)
			traceIDHex := "80f198ee56343ba864fe8b2a57d3eff7"
			spanIDHex := "e457b5a2e4d86bd1"
			sampled := true

			sc, err := createSpanContext(traceIDHex, spanIDHex, sampled)
			Expect(err).ToNot(HaveOccurred())

			injector := InjectHTTPHeaders(mockRequest)
			err = injector(sc)

			Expect(err).ToNot(HaveOccurred())
			value, ok := mockRequest.Header().Get(b3.TraceID)
			Expect(value).To(Equal(traceIDHex))
			Expect(ok).To(BeTrue())

			value, ok = mockRequest.Header().Get(b3.SpanID)
			Expect(value).To(Equal(spanIDHex))
			Expect(ok).To(BeTrue())

			value, ok = mockRequest.Header().Get(b3.Sampled)
			Expect(value).To(Equal("1"))
			Expect(ok).To(BeTrue())
		})

		It("should not inject headers when span context is empty", func() {
			mockRequest := newMockRequestWithHeaders(nil)

			sc := model.SpanContext{}

			injector := InjectHTTPHeaders(mockRequest)
			err := injector(sc)

			Expect(err.Error()).To(Equal("empty request context"))

			value, ok := mockRequest.Header().Get(b3.TraceID)
			Expect(value).To(BeEmpty())
			Expect(ok).To(BeFalse())

			value, ok = mockRequest.Header().Get(b3.SpanID)
			Expect(value).To(BeEmpty())
			Expect(ok).To(BeFalse())

			value, ok = mockRequest.Header().Get(b3.Sampled)
			Expect(value).To(BeEmpty())
			Expect(ok).To(BeFalse())
		})
	})
})

var _ = Describe("ExtractRemoteEndpoint", func() {
	It("should correctly extract the remote endpoint information", func() {
		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		req.RemoteAddr = "192.168.1.10:12345"
		wrappedReq := wrapper.WrapHTTPRequest(req)

		remoteEndpoint, err := ExtractRemoteEndpoint(wrappedReq)

		Expect(err).NotTo(HaveOccurred())
		Expect(remoteEndpoint).To(Equal(&model.Endpoint{
			ServiceName: "example.com",
			IPv4:        net.ParseIP("192.168.1.10"),
			Port:        12345,
		}))
	})

	It("should return an error if the remote address is invalid", func() {
		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		req.RemoteAddr = "invalid:address"
		wrappedReq := wrapper.WrapHTTPRequest(req)

		_, err := ExtractRemoteEndpoint(wrappedReq)

		Expect(err).To(HaveOccurred())
	})
})
