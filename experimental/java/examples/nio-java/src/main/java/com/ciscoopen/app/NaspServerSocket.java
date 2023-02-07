package com.ciscoopen.app;

import java.io.IOException;
import java.net.*;
import java.nio.channels.spi.SelectorProvider;

import nasp.Nasp;
import nasp.NaspIntegrationHandler;
import nasp.TCPListener;
import nasp.Connection;

public class NaspServerSocket extends ServerSocket {

    private final NaspIntegrationHandler NaspIntegrationHandler;
    private TCPListener TCPListener;
    private final SelectorProvider selectorProvider;

    private int localPort;
    private String localAddress;

    public NaspServerSocket(SelectorProvider selectorProvider) throws IOException {
        try {
            NaspIntegrationHandler = Nasp.newNaspIntegrationHandler("https://localhost:16443/config",
                    "test-tcp-16362813-F46B-41AC-B191-A390DB1F6BDF",
                    "16362813-F46B-41AC-B191-A390DB1F6BDF");

        } catch (Exception e) {
            throw new IOException("could not get nasp tcp listener");
        }

        this.selectorProvider = selectorProvider;
    }

    @Override
    public void bind(SocketAddress endpoint) throws IOException {
        if (endpoint instanceof InetSocketAddress) {
            try {
                localAddress = ((InetSocketAddress) endpoint).getHostString();
                localPort = ((InetSocketAddress) endpoint).getPort();
                TCPListener = NaspIntegrationHandler.bind(localAddress, localPort);
            } catch (Exception e) {
                throw new IOException(e);
            }
        } else {
            throw new UnsupportedOperationException();
        }
    }

    @Override
    public void bind(SocketAddress endpoint, int backlog) throws IOException {
        bind(endpoint);
    }
    @Override
    public int getLocalPort() {
        if (TCPListener != null) {
            return localPort;
        }
        return -1;
    }

    @Override
    public void setReuseAddress(boolean on) throws SocketException {

    }

    @Override
    public Socket accept() throws IOException {
        try {
            Connection conn = TCPListener.asyncAccept();
            return new NaspSocket(selectorProvider, conn);
        } catch (Exception e) {
            throw new IOException("could not bound to nasp tcp listener");
        }
    }

    public nasp.TCPListener getTCPListener() {
        return TCPListener;
    }
}
