package com.ciscoopen.app;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.Socket;
import java.net.SocketAddress;
import java.net.SocketException;
import java.nio.channels.SocketChannel;
import java.nio.channels.spi.SelectorProvider;

import nasp.Connection;

public class NaspSocket extends Socket {
    private final Connection conn;
    private final SocketChannel channel;

    public NaspSocket(SelectorProvider provider, Connection conn) {
        this.conn = conn;
        this.channel = new NaspSocketChannel(provider, conn, this);
    }
    @Override
    public int getPort() {
        try {
            return conn.getAddress().getPort();
        } catch (Exception e) {
            e.printStackTrace();
        }
        return 0;
    }


    @Override
    public InputStream getInputStream() throws IOException {
        return new NaspSocketInputStream(this.conn);
    }

    @Override
    public OutputStream getOutputStream() throws IOException {
        return new NaspSocketOutputStream(this.conn);
    }

    @Override
    public void bind(SocketAddress bindPoint) throws IOException {
        throw new UnsupportedOperationException();
    }

    public void setReuseAddress(boolean on) throws SocketException {
        throw new UnsupportedOperationException();
    }

    @Override
    public SocketChannel getChannel() {
        return channel;
    }
}
