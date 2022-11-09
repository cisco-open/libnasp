package com.ciscoopen.nasp;

import istio.HTTPTransport;
import istio.Istio;
import org.springframework.boot.web.server.WebServer;
import org.springframework.boot.web.server.WebServerException;
import org.springframework.http.server.reactive.HttpHandler;

public class NaspWebServer implements WebServer {
    private HTTPTransport transport;
    private final int port;
    private final NaspHttpHandler httpHandler;

    public NaspWebServer(int port, HttpHandler httpHandler) {
        this.port = port;
        this.httpHandler = new NaspHttpHandler(httpHandler);
    }

    @Override
    public void start() throws WebServerException {
        try {
            transport = Istio.newHTTPTransport("https://localhost:16443/config", "test-http-16362813-F46B-41AC-B191-A390DB1F6BDF", "16362813-F46B-41AC-B191-A390DB1F6BDF");
        } catch (Exception e) {
            throw new WebServerException("failed to create nasp transport", e);
        }

        try {
            transport.listenAndServe(":" + port, httpHandler);
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
