package com.ciscoopen.nasp;

import org.springframework.boot.web.reactive.server.AbstractReactiveWebServerFactory;
import org.springframework.boot.web.server.WebServer;
import org.springframework.http.server.reactive.HttpHandler;

public class NaspWebServerFactory extends AbstractReactiveWebServerFactory {

    @Override
    public WebServer getWebServer(HttpHandler httpHandler) {
        return new NaspWebServer(getPort(), httpHandler);
    }
}
