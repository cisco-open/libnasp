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
import java.nio.channels.Selector;
import java.nio.channels.SocketChannel;
import java.nio.channels.spi.SelectorProvider;
import java.util.Objects;
import java.util.Set;

public class NaspSocketChannel extends SocketChannel implements SelChImpl {

    private TCPDialer naspTcpDialer;
    private Connection connection;
    private boolean connected;
    private NaspSocket socket;
    private InetSocketAddress address;
    private Selector selector;

    public NaspSocketChannel(SelectorProvider provider) {
        super(provider);
        try {
            naspTcpDialer = Nasp.newTCPDialer();
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    protected NaspSocketChannel(SelectorProvider provider, Connection connection, NaspSocket socket) {
        super(provider);
        this.connection = connection;
        if (connection != null) {
            connected = true;
        }
        this.socket = socket;
        try {
            naspTcpDialer = Nasp.newTCPDialer();
        } catch (Exception e) {
            e.printStackTrace();
        }
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
        return connected;
    }

    @Override
    public boolean isConnectionPending() {
        return !connected;
    }

    @Override
    public boolean connect(SocketAddress remote) throws IOException {
        address = (InetSocketAddress) checkRemote(remote);
        socket.setAddress(address);
        if (selector != null) {
            naspTcpDialer.startAsyncDial(this.keyFor(selector).hashCode(), ((NaspSelector)selector).getSelector(),
                    address.getHostString(), address.getPort());
        }
        return false;
    }

    private SocketAddress checkRemote(SocketAddress sa) {
        String addr = "";
        try {
            addr = Nasp.checkAddress(
                    ((InetSocketAddress) sa).getHostString() );
        } catch (Exception e) {
            e.printStackTrace();
        }
        if (!addr.equals("")) {
            return new InetSocketAddress(addr, ((InetSocketAddress) sa).getPort());
        }
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
        if (socket.getConnection() == null) {
            socket.setConnection(connection);
            connected = true;
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
            int num = nativeRead(buff);
            if (num == -1) {
                return -1;
            }
            dst.put(buff, 0, num);
            return num;
        } catch (Exception e) {
            throw new IOException(e);
        }
    }

    private int nativeRead(byte[] buff) throws Exception {
        return connection.asyncRead(buff);
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
            int writtenBytes = nativeWrite(temp);

            if (writtenBytes < rem) {
                src.position(srcBufPos + writtenBytes);
            }
            return writtenBytes;
        } catch (Exception e) {
            throw new IOException(e);
        }
    }

    private int nativeWrite(byte[] temp) throws Exception {
        return connection.asyncWrite(temp);
    }

    @Override
    public long write(ByteBuffer[] srcs, int offset, int length) throws IOException {
        Objects.checkFromIndexSize(offset, length, srcs.length);

        int totalLength = 0;
        for (int i = offset; i < offset + length; i++) {
            int expectedWriteCount =  srcs[i].remaining();
            if (expectedWriteCount == 0) {
                continue;
            }
            int writtenBytes = write(srcs[i]);
            totalLength += writtenBytes;

            if (writtenBytes < expectedWriteCount) {
                // in case the data of one of the byte buffers from the array can not be written to the connection
                // entirely we need to stop the write operation to ensure the byte buffers are written out in order
                break;
            }
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

    public void setSelector(Selector selector) {
        this.selector = selector;
    }
}
