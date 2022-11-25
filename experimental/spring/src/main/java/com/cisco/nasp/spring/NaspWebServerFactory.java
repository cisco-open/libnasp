package com.cisco.nasp.spring;

import org.springframework.boot.web.reactive.server.AbstractReactiveWebServerFactory;
import org.springframework.boot.web.server.WebServer;
import org.springframework.http.server.reactive.HttpHandler;

public class NaspWebServerFactory extends AbstractReactiveWebServerFactory {

    private final NaspConfiguration configuration;

    public NaspWebServerFactory(NaspConfiguration configuration) {
        this.configuration = configuration;
    }

    @Override
    public WebServer getWebServer(HttpHandler httpHandler) {
        return new NaspWebServer(configuration, getPort(), httpHandler);
    }
}
