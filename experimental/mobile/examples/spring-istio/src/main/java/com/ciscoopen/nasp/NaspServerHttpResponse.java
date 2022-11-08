package com.ciscoopen.nasp;

import istio.NaspResponseWriter;
import org.reactivestreams.Publisher;
import org.springframework.core.io.buffer.DataBuffer;
import org.springframework.core.io.buffer.DefaultDataBufferFactory;
import org.springframework.http.server.reactive.AbstractServerHttpResponse;
import reactor.core.publisher.Mono;

public class NaspServerHttpResponse extends AbstractServerHttpResponse {

    private final NaspResponseWriter responseWriter;

    public NaspServerHttpResponse(istio.NaspResponseWriter responseWriter) {
        super(DefaultDataBufferFactory.sharedInstance);
        this.responseWriter = responseWriter;
    }

    @Override
    public <T> T getNativeResponse() {
        return (T) responseWriter;
    }

    @Override
    protected Mono<Void> writeWithInternal(Publisher<? extends DataBuffer> body) {
        return Mono.from(body).map(buffer -> {
            try {
                byte[] bytes = buffer.asInputStream().readNBytes(buffer.readableByteCount());
                return responseWriter.write(bytes);
            } catch (Exception e) {
                throw new RuntimeException(e);
            }
        }).then();
    }

    @Override
    protected Mono<Void> writeAndFlushWithInternal(Publisher<? extends Publisher<? extends DataBuffer>> body) {
        return null;
    }

    @Override
    protected void applyStatusCode() {

    }

    @Override
    protected void applyHeaders() {

    }

    @Override
    protected void applyCookies() {

    }
}
