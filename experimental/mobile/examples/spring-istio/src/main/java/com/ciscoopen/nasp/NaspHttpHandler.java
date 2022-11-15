package com.ciscoopen.nasp;

import istio.NaspHttpRequest;
import istio.NaspResponseWriter;
import org.springframework.http.server.reactive.HttpHandler;

public class NaspHttpHandler implements istio.HttpHandler {

    private final HttpHandler httpHandler;

    public NaspHttpHandler(HttpHandler httpHandler) {
        this.httpHandler = httpHandler;
    }

    @Override
    public void serveHTTP(NaspResponseWriter naspResponse, NaspHttpRequest naspRequest) {
        NaspServerHttpRequest request = new NaspServerHttpRequest(naspRequest);
        NaspServerHttpResponse response = new NaspServerHttpResponse(naspResponse);
        httpHandler.handle(request, response).block();
    }
}
