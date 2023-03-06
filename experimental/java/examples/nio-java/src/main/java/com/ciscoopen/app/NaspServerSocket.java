package com.ciscoopen.app;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.ServerSocket;
import java.net.Socket;
import java.net.SocketAddress;
import java.net.SocketException;
import java.nio.channels.spi.SelectorProvider;

import nasp.Nasp;
import nasp.TCPListener;
import nasp.Connection;

public class NaspServerSocket extends ServerSocket {

    private TCPListener naspTcpListener;
    private final SelectorProvider selectorProvider;

    private int localPort;
    private InetSocketAddress address;

    public NaspServerSocket(SelectorProvider selectorProvider) throws IOException {
        this.selectorProvider = selectorProvider;
    }

    @Override
    public void bind(SocketAddress endpoint) throws IOException {
        if (endpoint instanceof InetSocketAddress) {
            try {
                address = (InetSocketAddress) endpoint;
                localPort = ((InetSocketAddress) endpoint).getPort();
                naspTcpListener = Nasp.bind(((InetSocketAddress) endpoint).getHostString(), localPort);
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
        if (naspTcpListener != null) {
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
            Connection conn = naspTcpListener.asyncAccept();
            if (conn == null) {
                return null;
            }
            return new NaspSocket(selectorProvider, conn);
        } catch (Exception e) {
            throw new IOException("could not bound to nasp tcp listener");
        }
    }

    public TCPListener getNaspTcpListener() {
        return naspTcpListener;
    }
}
