package com.ciscoopen.nasp;

import istio.NaspHttpRequest;
import istio.NaspResponseWriter;
import org.springframework.http.server.reactive.HttpHandler;
import reactor.core.publisher.Mono;

public class NaspHttpHandler implements istio.HttpHandler {

    private final HttpHandler httpHandler;

    public NaspHttpHandler(HttpHandler httpHandler) {
        this.httpHandler = httpHandler;
    }

    @Override
    public void serveHTTP(NaspResponseWriter naspResponse, NaspHttpRequest naspRequest) {
        NaspServerHttpRequest request = new NaspServerHttpRequest(naspRequest);
        NaspServerHttpResponse response = new NaspServerHttpResponse(naspResponse);
        Mono<Void> handle = httpHandler.handle(request, response);
        handle.block();
    }
}
