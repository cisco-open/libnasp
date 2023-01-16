package com.ciscoopen.app;

import java.io.IOException;
import java.net.SocketAddress;
import java.net.SocketOption;
import java.nio.channels.ServerSocketChannel;
import java.nio.channels.SocketChannel;
import java.nio.channels.spi.SelectorProvider;
import java.util.Set;

public class NaspServerSocketChannel extends ServerSocketChannel {
    private WrappedServerSocket socket;

    public NaspServerSocketChannel(SelectorProvider sp) {
        super(sp);
    }

    @Override
    public ServerSocketChannel bind(SocketAddress local, int backlog) throws IOException {
        return null;
    }

    @Override
    public <T> ServerSocketChannel setOption(SocketOption<T> name, T value) throws IOException {
        return null;
    }

    @Override
    public <T> T getOption(SocketOption<T> name) throws IOException {
        return null;
    }

    @Override
    public Set<SocketOption<?>> supportedOptions() {
        return null;
    }

    public WrappedServerSocket socket() {
        try {
            if (this.socket == null) {
                this.socket = new WrappedServerSocket(10000);
            }
            return socket;
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }

//    public final SelectableChannel configureBlocking(boolean block) throws IOException {
//        return this.channel.configureBlocking(block);
//    }

    @Override
    protected void implConfigureBlocking(boolean block) throws IOException {

    }

//    public final SelectionKey register(Selector sel, int ops, Object att) throws ClosedChannelException {
//        return this.channel.register(sel, ops, att);
//    }

    @Override
    protected void implCloseSelectableChannel() throws IOException {

    }

//    public final SelectionKey register(Selector sel, int ops)
//            throws ClosedChannelException
//    {
//        return register(sel, ops, null);
//    }

    public SocketChannel accept() throws IOException {
        return this.socket.accept().getChannel();
    }

    @Override
    public SocketAddress getLocalAddress() throws IOException {
        return null;
    }

}
