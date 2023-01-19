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

public class WrappedSocket extends Socket {
    private final Connection conn;
    private final SocketChannel channel;

    public WrappedSocket(SelectorProvider provider, Connection conn) {
        this.conn = conn;
        this.channel = new NaspSocketChannel(provider, conn, this);
    }
    @Override
    public int getPort() {
        return 10000; //TODO fix this
    }


    @Override
    public InputStream getInputStream() throws IOException {
        return new WrappedInputStream(this.conn);
    }

    @Override
    public OutputStream getOutputStream() throws IOException {
        return new WrappedOutputStream(this.conn);
    }

    @Override
    public void bind(SocketAddress bindPoint) throws IOException {

    }

    public void setReuseAddress(boolean on) throws SocketException {

    }

    @Override
    public SocketChannel getChannel() {
        return channel;
    }
}
