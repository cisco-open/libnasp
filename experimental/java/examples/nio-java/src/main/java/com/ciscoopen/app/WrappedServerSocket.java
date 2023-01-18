package com.ciscoopen.app;

import java.io.IOException;
import java.net.ServerSocket;
import java.net.Socket;
import java.net.SocketAddress;
import java.net.SocketException;
import java.nio.channels.spi.SelectorProvider;

import nasp.Nasp;
import nasp.TCPListener;
import nasp.Connection;

public class WrappedServerSocket extends ServerSocket {
    TCPListener TCPListener;
    SelectorProvider selectorProvider;

    public WrappedServerSocket(int port, SelectorProvider selectorProvider) throws IOException {
        try {
            this.TCPListener = Nasp.newTCPListener("https://localhost:16443/config",
                    "test-tcp-16362813-F46B-41AC-B191-A390DB1F6BDF",
                    "16362813-F46B-41AC-B191-A390DB1F6BDF");

        } catch (Exception e) {
            throw new IOException("could not get nasp tcp listener");
        }

        this.selectorProvider = selectorProvider;
    }
    @Override
    public void bind(SocketAddress endpoint) throws IOException {
    }

    @Override
    public void setReuseAddress(boolean on) throws SocketException {

    }

    @Override
    public Socket accept() throws IOException {
        try {
            Connection conn = this.TCPListener.asyncAccept();
            return new WrappedSocket(selectorProvider, conn);
        } catch (Exception e) {
            throw new IOException("could not bound to nasp tcp listener");
        }
    }

}
