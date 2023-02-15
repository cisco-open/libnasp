package com.ciscoopen.app;

import nasp.Connection;
import nasp.TCPDialer;
import nasp.Nasp;
import sun.nio.ch.Net;
import sun.nio.ch.SelChImpl;
import sun.nio.ch.SelectionKeyImpl;

import java.io.FileDescriptor;
import java.io.IOException;
import java.lang.reflect.Constructor;
import java.lang.reflect.InvocationTargetException;
import java.net.*;
import java.nio.ByteBuffer;
import java.nio.channels.SelectionKey;
import java.nio.channels.SocketChannel;
import java.nio.channels.spi.SelectorProvider;
import java.util.ArrayList;
import java.util.Objects;
import java.util.Set;

import static java.net.StandardProtocolFamily.INET6;

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
        if (socket == null) {
            socket = new NaspSocket(provider(), connection);
        }
        return socket;
    }

    @Override
    public boolean isConnected() {
        return connection != null;
    }

    @Override
    public boolean isConnectionPending() {
        return connection == null;
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
        address = (InetSocketAddress) checkRemote(remote);
        socket.setAddress(address);
        return false;
    }

    private SocketAddress checkRemote(SocketAddress sa) {
        InetSocketAddress isa = Net.checkAddress(sa);
        @SuppressWarnings("removal")
        SecurityManager sm = System.getSecurityManager();
        if (sm != null) {
            sm.checkConnect(isa.getAddress().getHostAddress(), isa.getPort());
        }
        InetAddress address = isa.getAddress();
        if (address.isAnyLocalAddress()) {
            int port = isa.getPort();
            if (address instanceof Inet4Address) {
                return new InetSocketAddress("127.0.0.1", port);
            } else {
                return new InetSocketAddress("::1", port);
            }
        } else {
            return isa;
        }
    }

    @Override
    public boolean finishConnect() throws IOException {
        connection = dialer.asyncDial();
        if (connection == null) {
            return false;
        }
        return true;
    }

    @Override
    public SocketAddress getRemoteAddress() throws IOException {
        return address;
    }

    @Override
    public int read(ByteBuffer dst) throws IOException {
        //TODO revise this to be more efficient
        try {
            byte[] buff = new byte[dst.remaining()];
            int num = connection.asyncRead(buff);
            dst.put(buff, dst.position(), num);
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
            int rem = src.remaining();
            if (rem == 0) {
                return 0;
            }

            byte[] temp = new byte[rem];
            src.get(temp, src.position(), src.limit());
            return connection.asyncWrite(temp);
        } catch (Exception e) {
            throw new IOException(e);
        }
    }

    @Override
    public long write(ByteBuffer[] srcs, int offset, int length) throws IOException {
        Objects.checkFromIndexSize(offset, length, srcs.length);

        try {
            int totalLength = 0;
            for (int i = offset; i < offset + length; i++) {
                ByteBuffer src = srcs[i];
                totalLength += write(src);
            }
            return totalLength;
        } catch (IOException e) {
            throw e;
        }
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
        connection.close();
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
