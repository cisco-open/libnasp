package com.ciscoopen.app;

import sun.nio.ch.Net;
import sun.nio.ch.SelChImpl;
import sun.nio.ch.SelectionKeyImpl;

import java.io.FileDescriptor;
import java.io.IOException;
import java.net.Socket;
import java.net.SocketAddress;
import java.net.SocketOption;
import java.nio.channels.SelectionKey;
import java.nio.channels.ServerSocketChannel;
import java.nio.channels.SocketChannel;
import java.nio.channels.spi.SelectorProvider;
import java.util.Set;

public class NaspServerSocketChannel extends ServerSocketChannel implements SelChImpl {
    private NaspServerSocket socket;

    public NaspServerSocketChannel(SelectorProvider sp) {
        super(sp);
    }

    @Override
    public ServerSocketChannel bind(SocketAddress local, int backlog) throws IOException {
        throw new UnsupportedOperationException();
    }

    @Override
    public <T> ServerSocketChannel setOption(SocketOption<T> name, T value) throws IOException {
        throw new UnsupportedOperationException();
    }

    @Override
    public <T> T getOption(SocketOption<T> name) throws IOException {
        throw new UnsupportedOperationException();
    }

    @Override
    public Set<SocketOption<?>> supportedOptions() {
        throw new UnsupportedOperationException();
    }

    public NaspServerSocket socket() {
        try {
            if (this.socket == null) {
                this.socket = new NaspServerSocket(provider());
            }
            return socket;
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }

    @Override
    protected void implConfigureBlocking(boolean block) throws IOException {

    }

    @Override
    protected void implCloseSelectableChannel() throws IOException {
        socket.getNaspTcpListener().close();
    }

    public SocketChannel accept() throws IOException {
        Socket client = socket.accept();
        if (client == null) {
            return null;
        }
        return client.getChannel();
    }

    @Override
    public SocketAddress getLocalAddress() throws IOException {
        throw new UnsupportedOperationException();
    }

    @Override
    public FileDescriptor getFD() {
        throw new UnsupportedOperationException();
    }

    @Override
    public int getFDVal() {
        throw new UnsupportedOperationException();
    }

    public boolean translateAndUpdateReadyOps(int ops, SelectionKeyImpl ski) {
        return translateReadyOps(ops, ski.nioReadyOps(), ski);
    }

    public boolean translateAndSetReadyOps(int ops, SelectionKeyImpl ski) {
        return translateReadyOps(ops, 0, ski);
    }

    /**
     * Translates native poll revent set into a ready operation set
     */
    public boolean translateReadyOps(int ops, int initialOps, SelectionKeyImpl ski) {
        int intOps = ski.nioInterestOps();
        int oldOps = ski.nioReadyOps();
        int newOps = initialOps;

        if ((ops & Net.POLLNVAL) != 0) {
            // This should only happen if this channel is pre-closed while a
            // selection operation is in progress
            // ## Throw an error if this channel has not been pre-closed
            return false;
        }

        if ((ops & (Net.POLLERR | Net.POLLHUP)) != 0) {
            newOps = intOps;
            ski.nioReadyOps(newOps);
            return (newOps & ~oldOps) != 0;
        }

        if (((ops & Net.POLLIN) != 0) &&
                ((intOps & SelectionKey.OP_ACCEPT) != 0))
            newOps |= SelectionKey.OP_ACCEPT;

        ski.nioReadyOps(newOps);
        return (newOps & ~oldOps) != 0;
    }

    /**
     * Translates an interest operation set into a native poll event set
     */
    public int translateInterestOps(int ops) {
        int newOps = 0;
        if ((ops & SelectionKey.OP_ACCEPT) != 0)
            newOps |= Net.POLLIN;
        return newOps;
    }

    @Override
    public void kill() throws IOException {
        socket.getNaspTcpListener().close();
    }
}
