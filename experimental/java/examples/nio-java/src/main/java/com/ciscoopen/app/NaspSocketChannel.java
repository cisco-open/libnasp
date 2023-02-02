package com.ciscoopen.app;

import nasp.Connection;
import nasp.TCPDialer;
import nasp.Nasp;
import sun.nio.ch.Net;
import sun.nio.ch.SelChImpl;
import sun.nio.ch.SelectionKeyImpl;

import java.io.FileDescriptor;
import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.SocketAddress;
import java.net.SocketOption;
import java.nio.ByteBuffer;
import java.nio.channels.SelectionKey;
import java.nio.channels.SocketChannel;
import java.nio.channels.spi.SelectorProvider;
import java.util.Set;

public class NaspSocketChannel extends SocketChannel implements SelChImpl {

    private TCPDialer dialer;
    private Connection connection;
    private NaspSocket socket;
    private InetSocketAddress address;

    public NaspSocketChannel(SelectorProvider provider) {
        super(provider);
    }

    protected NaspSocketChannel(SelectorProvider provider, Connection connection, NaspSocket socket) {
        super(provider);
        this.connection = connection;
        this.socket = socket;
    }

    @Override
    public SocketChannel bind(SocketAddress local) throws IOException {
        throw new UnsupportedOperationException();
    }

    @Override
    public <T> SocketChannel setOption(SocketOption<T> name, T value) throws IOException {
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

    @Override
    public SocketChannel shutdownInput() throws IOException {
        throw new UnsupportedOperationException();
    }

    @Override
    public SocketChannel shutdownOutput() throws IOException {
        throw new UnsupportedOperationException();
    }

    @Override
    public NaspSocket socket() {
        return socket;
    }

    @Override
    public boolean isConnected() {
        //TODO check this if this is right the whole time
        return true;
    }

    @Override
    public boolean isConnectionPending() {
        return false;
    }

    @Override
    public boolean connect(SocketAddress remote) throws IOException {
        try {
            dialer = Nasp.newTCPDialer("https://localhost:16443/config",
                    "test-tcp-16362813-F46B-41AC-B191-A390DB1F6BDF",
                    "16362813-F46B-41AC-B191-A390DB1F6BDF");
        } catch (Exception e) {
            throw new IOException("could not get nasp tcp dialer");
        }
        address = (InetSocketAddress)remote;
        return false;
    }

    @Override
    public boolean finishConnect() throws IOException {
        connection = dialer.asyncDial();
        return true;
    }

    @Override
    public SocketAddress getRemoteAddress() throws IOException {
        throw new UnsupportedOperationException();
    }

    @Override
    public int read(ByteBuffer dst) throws IOException {
        try {
            byte[] buff = new byte[dst.remaining()];
            int num = connection.asyncRead(buff);
            dst.put(buff, 0, num);
            return num;
        } catch (Exception e) {
            throw new IOException(e);
        }
    }

    @Override
    public long read(ByteBuffer[] dsts, int offset, int length) throws IOException {
        throw new UnsupportedOperationException();
    }

    @Override
    public int write(ByteBuffer src) throws IOException {
        try {
            byte[] tempArray = new byte[src.limit()];
            src.get(tempArray, 0, src.limit());
            return connection.asyncWrite(tempArray);
        } catch (Exception e) {
            throw new IOException(e);
        }
    }

    @Override
    public long write(ByteBuffer[] srcs, int offset, int length) throws IOException {
        throw new UnsupportedOperationException();
    }

    @Override
    public SocketAddress getLocalAddress() throws IOException {
        throw new UnsupportedOperationException();
    }

    @Override
    protected void implCloseSelectableChannel() throws IOException {
        connection.close();
    }

    @Override
    protected void implConfigureBlocking(boolean block) throws IOException {

    }

    @Override
    public FileDescriptor getFD() {

        throw new UnsupportedOperationException();
    }

    @Override
    public int getFDVal() {
        throw new UnsupportedOperationException();
    }

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

        boolean connected = isConnected();
        if (((ops & Net.POLLIN) != 0) &&
                ((intOps & SelectionKey.OP_READ) != 0) && connected)
            newOps |= SelectionKey.OP_READ;

        if (((ops & Net.POLLCONN) != 0) &&
                ((intOps & SelectionKey.OP_CONNECT) != 0) && isConnectionPending())
            newOps |= SelectionKey.OP_CONNECT;

        if (((ops & Net.POLLOUT) != 0) &&
                ((intOps & SelectionKey.OP_WRITE) != 0) && connected)
            newOps |= SelectionKey.OP_WRITE;

        ski.nioReadyOps(newOps);
        return (newOps & ~oldOps) != 0;
    }

    @Override
    public boolean translateAndUpdateReadyOps(int ops, SelectionKeyImpl ski) {
        return translateReadyOps(ops, ski.nioReadyOps(), ski);
    }

    @Override
    public boolean translateAndSetReadyOps(int ops, SelectionKeyImpl ski) {
        return translateReadyOps(ops, 0, ski);
    }

    @Override
    public int translateInterestOps(int ops) {
        int newOps = 0;
        if ((ops & SelectionKey.OP_READ) != 0)
            newOps |= Net.POLLIN;
        if ((ops & SelectionKey.OP_WRITE) != 0)
            newOps |= Net.POLLOUT;
        if ((ops & SelectionKey.OP_CONNECT) != 0)
            newOps |= Net.POLLCONN;
        return newOps;
    }

    @Override
    public void kill() throws IOException {
        throw new UnsupportedOperationException();
    }

    public Connection getConnection() {
        return connection;
    }

    public TCPDialer getTCPDialer() {
        return dialer;
    }

    public InetSocketAddress getAddress() {
        return address;
    }
}
