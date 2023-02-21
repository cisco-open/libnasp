package com.ciscoopen.app;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.*;
import java.nio.channels.SocketChannel;
import java.nio.channels.spi.SelectorProvider;

import nasp.Connection;

public class NaspSocket extends Socket {
    private final Connection conn;
    private final SocketChannel channel;

    private InetSocketAddress address;

    public NaspSocket(SelectorProvider provider, Connection conn) {
        this.conn = conn;
        this.channel = new NaspSocketChannel(provider, conn, this);
    }
    public void setAddress(InetSocketAddress address) {
        this.address = address;
    }
    @Override
    public int getPort() {
        try {
            return conn.getRemoteAddress().getPort();
        } catch (Exception e) {
            e.printStackTrace();
        }
        return -1;
    }

    @Override
    public int getLocalPort() {
        try {
            return conn.getAddress().getPort();
        } catch (Exception e) {
            e.printStackTrace();
        }
        return -1;
    }

    @Override
    public InetAddress getInetAddress() {
        if (conn != null) {
            try {
                return InetAddress.getByName(conn.getRemoteAddress().getHost());
            } catch (Exception e) {
                e.printStackTrace();
            }
        }
        return address.getAddress();
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
