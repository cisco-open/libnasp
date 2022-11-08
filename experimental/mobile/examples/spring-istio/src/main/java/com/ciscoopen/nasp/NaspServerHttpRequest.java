package com.ciscoopen.nasp;

import com.fasterxml.jackson.databind.ObjectMapper;
import istio.NaspHttpRequest;
import org.springframework.core.io.buffer.DataBuffer;
import org.springframework.http.HttpCookie;
import org.springframework.http.HttpHeaders;
import org.springframework.http.server.reactive.AbstractServerHttpRequest;
import org.springframework.http.server.reactive.SslInfo;
import org.springframework.util.MultiValueMap;
import reactor.core.publisher.Flux;

import java.io.IOException;
import java.net.URI;

public class NaspServerHttpRequest extends AbstractServerHttpRequest {
    private final NaspHttpRequest request;
    private static final ObjectMapper JSON = new ObjectMapper();

    static HttpHeaders parseHeaders(NaspHttpRequest request) {
        try {
            return JSON.readValue(request.headers(), HttpHeaders.class);
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }

    public NaspServerHttpRequest(NaspHttpRequest request) {
        super(URI.create("http://localhost:8080/"), "", parseHeaders(request));
        this.request = request;
    }

    @Override
    protected MultiValueMap<String, HttpCookie> initCookies() {
        return null;
    }

    @Override
    protected SslInfo initSslInfo() {
        return null;
    }

    @Override
    public <T> T getNativeRequest() {
        return (T) request;
    }

    @Override
    public String getMethodValue() {
        return request.method();
    }

    @Override
    public Flux<DataBuffer> getBody() {
        return null;
    }
}
