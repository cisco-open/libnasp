package com.cisco.nasp.spring;

import com.fasterxml.jackson.databind.ObjectMapper;
import nasp.NaspHttpRequest;
import org.springframework.core.io.buffer.DataBuffer;
import org.springframework.core.io.buffer.DataBufferUtils;
import org.springframework.core.io.buffer.DefaultDataBufferFactory;
import org.springframework.http.HttpCookie;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpMethod;
import org.springframework.http.server.reactive.AbstractServerHttpRequest;
import org.springframework.http.server.reactive.SslInfo;
import org.springframework.util.MultiValueMap;
import reactor.core.publisher.Flux;

import java.io.IOException;
import java.io.InputStream;
import java.net.URI;

public class NaspServerHttpRequest extends AbstractServerHttpRequest {
    private final NaspHttpRequest request;
    private static final ObjectMapper JSON = new ObjectMapper();

    static HttpHeaders parseHeaders(NaspHttpRequest request) {
        try {
            return JSON.readValue(request.header(), HttpHeaders.class);
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }

    public NaspServerHttpRequest(NaspHttpRequest request) {
        super(URI.create(request.uri()), "", parseHeaders(request));
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
    public HttpMethod getMethod() {
        return HttpMethod.valueOf(request.method());
    }

    @Override
    public Flux<DataBuffer> getBody() {
        return DataBufferUtils.readInputStream(() -> new InputStream() {
            final nasp.Body body = request.body();
            final byte[] buffer = new byte[1];

            @Override
            public int read() throws IOException {
                try {
                    body.read(buffer);
                    return buffer[0];
                } catch (Exception e) {
                    if ("EOF".equals(e.getMessage())) {
                        return -1;
                    }
                    throw new IOException(e);
                }
            }

            @Override
            public void close() throws IOException {
                try {
                    body.close();
                } catch (Exception e) {
                    throw new IOException(e);
                }
            }
        }, DefaultDataBufferFactory.sharedInstance, 512);
    }
}
