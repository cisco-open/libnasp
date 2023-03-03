package com.ciscoopen.app;

import nasp.Connection;
import nasp.Nasp;
import nasp.TCPDialer;
import sun.nio.ch.Net;
import sun.nio.ch.SelChImpl;
import sun.nio.ch.SelectionKeyImpl;

import java.io.FileDescriptor;
import java.io.IOException;
import java.net.Inet4Address;
import java.net.InetAddress;
import java.net.InetSocketAddress;
import java.net.SocketAddress;
import java.net.SocketOption;
import java.nio.ByteBuffer;
import java.nio.channels.SelectionKey;
import java.nio.channels.SocketChannel;
import java.nio.channels.spi.SelectorProvider;
import java.util.Objects;
import java.util.Set;

public class NaspSocketChannel extends SocketChannel implements SelChImpl {

    private TCPDialer naspTcpDialer;
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
        return this;
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
            naspTcpDialer = Nasp.newTCPDialer("https://localhost:16443/config", System.getenv("NASP_AUTH_TOKEN"));
        } catch (Exception e) {
            throw new IOException("could not get nasp tcp dialer", e);
        }
        address = (InetSocketAddress) checkRemote(remote);
        socket.setAddress(address);
        return false;
    }

    private SocketAddress checkRemote(SocketAddress sa) {
        InetSocketAddress isa = Net.checkAddress(sa);
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
        connection = naspTcpDialer.asyncDial();
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
        try {
            byte[] buff = new byte[dst.remaining()];
            int num = connection.asyncRead(buff);
            if (num == -1) {
                return -1;
            }
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
            int rem = src.remaining();
            if (rem == 0) {
                return 0;
            }

            int srcBufPos = src.position();
            byte[] temp = new byte[rem];
            src.get(temp, 0, rem);
            int writtenBytes = connection.asyncWrite(temp);

            if (writtenBytes < rem) {
                src.position(srcBufPos + writtenBytes);
            }

            return writtenBytes;
        } catch (Exception e) {
            throw new IOException(e);
        }
    }

    @Override
    public long write(ByteBuffer[] srcs, int offset, int length) throws IOException {
        Objects.checkFromIndexSize(offset, length, srcs.length);

        int totalLength = 0;
        for (int i = offset; i < offset + length; i++) {
            int writtenBytes = write(srcs[i]);
            if (writtenBytes == -1) {
                return -1;
            }
            totalLength += writtenBytes;

        }
        return totalLength;
    }

    @Override
    public SocketAddress getLocalAddress() throws IOException {
        throw new UnsupportedOperationException();
    }

    @Override
    protected void implCloseSelectableChannel() throws IOException {
        if (connection != null) {
            connection.close();
        }
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
        if (((ops & Net.POLLIN) != 0) && ((intOps & SelectionKey.OP_READ) != 0) && connected)
            newOps |= SelectionKey.OP_READ;

        if (((ops & Net.POLLCONN) != 0) && ((intOps & SelectionKey.OP_CONNECT) != 0) && isConnectionPending())
            newOps |= SelectionKey.OP_CONNECT;

        if (((ops & Net.POLLOUT) != 0) && ((intOps & SelectionKey.OP_WRITE) != 0) && connected)
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
        if ((ops & SelectionKey.OP_READ) != 0) newOps |= Net.POLLIN;
        if ((ops & SelectionKey.OP_WRITE) != 0) newOps |= Net.POLLOUT;
        if ((ops & SelectionKey.OP_CONNECT) != 0) newOps |= Net.POLLCONN;
        return newOps;
    }

    @Override
    public void kill() throws IOException {
        if (connection != null) {
            connection.close();
        }
    }

    public Connection getConnection() {
        return connection;
    }

    public TCPDialer getNaspTcpDialer() {
        return naspTcpDialer;
    }

    public InetSocketAddress getAddress() {
        return address;
    }
}
