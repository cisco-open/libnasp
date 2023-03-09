package com.cisco.nasp.spring;

import nasp.HTTPTransport;
import nasp.Nasp;
import org.springframework.boot.web.server.WebServer;
import org.springframework.boot.web.server.WebServerException;
import org.springframework.http.server.reactive.HttpHandler;

public class NaspWebServer implements WebServer {
    private HTTPTransport transport;
    private final NaspConfiguration configuration;
    private final int port;
    private final NaspHttpHandler httpHandler;

    public NaspWebServer(NaspConfiguration configuration, int port, HttpHandler httpHandler) {
        this.configuration = configuration;
        this.port = port;
        this.httpHandler = new NaspHttpHandler(httpHandler);
    }

    @Override
    public void start() throws WebServerException {
        try {
            transport = Nasp.newHTTPTransport();
        } catch (Exception e) {
            throw new WebServerException("failed to create nasp transport", e);
        }

        try {
            Nasp.listenAndServe(":" + port, httpHandler);
        } catch (Exception e) {
            throw new WebServerException("failed to listen on nasp transport", e);
        }

        startDaemonAwaitThread(transport);
    }

    private void startDaemonAwaitThread(HTTPTransport transport) {
        Thread awaitThread = new Thread("server") {

            @Override
            public void run() {
                try {
                    transport.await();
                } catch (Exception e) {
                    throw new RuntimeException(e);
                }
            }
        };
        awaitThread.setContextClassLoader(getClass().getClassLoader());
        awaitThread.setDaemon(false);
        awaitThread.start();
    }

    @Override
    public void stop() throws WebServerException {
        if (transport != null) {
            transport.close();
        }
    }

    @Override
    public int getPort() {
        return port;
    }
}
