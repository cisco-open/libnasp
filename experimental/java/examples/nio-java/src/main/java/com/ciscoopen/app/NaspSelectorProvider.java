package com.ciscoopen.app;

import sun.nio.ch.FileChannelImpl;
import sun.nio.ch.SelectionKeyImpl;
import sun.nio.ch.SelectorImpl;
import sun.nio.ch.SelectorProviderImpl;

import java.io.IOException;
import java.lang.reflect.Field;
import java.net.InetSocketAddress;
import java.nio.channels.SelectionKey;
import java.nio.channels.Selector;
import java.nio.channels.ServerSocketChannel;
import java.nio.channels.SocketChannel;
import java.nio.channels.spi.AbstractSelector;
import java.nio.channels.spi.SelectorProvider;
import java.util.HashMap;
import java.util.Map;
import java.util.function.Consumer;

class NaspSelector extends SelectorImpl {

    static int getOperation(long selectedKey) {
        return (int) (selectedKey >> 32);
    }

    static int getSelectedKeyId(long selectedKey) {
        return (int) selectedKey;
    }

    private final nasp.Selector selector = new nasp.Selector();
    private final Map<Integer, SelectionKeyImpl> selectionKeyTable = new HashMap<>();
    private final Map<Integer, Integer> runningAsyncOps = new HashMap<>();

    protected NaspSelector(SelectorProvider sp) {
        super(sp);
    }

    @Override
    protected int doSelect(Consumer<SelectionKey> action, long timeout) throws IOException {
        selector.select(timeout);

        int numKeysUpdated = 0;
        long selectedKey;
        while ((selectedKey = selector.nextSelectedKey()) != 0) {
            SelectionKeyImpl selectionKey = selectionKeyTable.get(getSelectedKeyId(selectedKey));
            if (selectionKey != null) {
                if (selectionKey.isValid()) {
                    numKeysUpdated += processReadyEvents(getOperation(selectedKey), selectionKey, action);
                }
            }
        }

        return numKeysUpdated;
    }

    @Override
    public Selector wakeup() {
        selector.wakeUp();
        return this;
    }

    @Override
    protected void implRegister(SelectionKeyImpl ski) {
        super.implRegister(ski);
        selectionKeyTable.put(ski.hashCode(), ski);

        int interestOps = ski.interestOps();
        startAsyncOps(ski, interestOps);
    }

    @Override
    protected void implClose() throws IOException {
        selector.close();
    }

    @Override
    protected void implDereg(SelectionKeyImpl ski) throws IOException {
        throw new UnsupportedOperationException();
    }

    @Override
    protected void setEventOps(SelectionKeyImpl ski) {
        int interestOps = newInterestOps(ski);

        startAsyncOps(ski, interestOps);
    }

    protected int newInterestOps(SelectionKeyImpl ski) {
        if (ski == null)
            return 0;

        int interestOps = ski.interestOps();
        Integer runningOps = runningAsyncOps.get(ski.hashCode());
        if (runningOps == null) {
            return interestOps;
        }

        int mask = interestOps ^ runningOps;

        return interestOps & mask;
    }

    protected void startAsyncOps(SelectionKeyImpl ski, int interestOps) {
        int selectedKeyId = ski.hashCode();
        int runningOps = 0;
        if (runningAsyncOps.containsKey(selectedKeyId)) {
            runningOps = runningAsyncOps.get(selectedKeyId);
        }

        if (ski.channel() instanceof NaspServerSocketChannel naspServerSockChan) {
            startAsyncOps(naspServerSockChan, ski.hashCode(), runningOps, interestOps);
            return;
        }

        if (ski.channel() instanceof NaspSocketChannel naspSockChan) {
            startAsyncOps(naspSockChan, ski.hashCode(), runningOps, interestOps);
        }
    }

    protected void startAsyncOps(NaspSocketChannel naspSocketChannel, int selectedKeyId, int runningOps, int interestOps) {
        if ((interestOps & SelectionKey.OP_READ) != 0) {
            naspSocketChannel.getConnection().startAsyncRead(selectedKeyId, selector);
            runningAsyncOps.put(selectedKeyId, runningOps | SelectionKey.OP_READ);
        }
        if ((interestOps & SelectionKey.OP_WRITE) != 0) {
            naspSocketChannel.getConnection().startAsyncWrite(selectedKeyId, selector);
            runningAsyncOps.put(selectedKeyId, runningOps | SelectionKey.OP_WRITE);
        }
        if ((interestOps & SelectionKey.OP_CONNECT) != 0) {
            //This could happen if we are servers not clients
            if (naspSocketChannel.getNaspTcpDialer() != null) {
                InetSocketAddress address = naspSocketChannel.getAddress();
                naspSocketChannel.getNaspTcpDialer().startAsyncDial(selectedKeyId, selector,
                        address.getHostString(), address.getPort());
                runningAsyncOps.put(selectedKeyId, runningOps | SelectionKey.OP_CONNECT);
            }
        }
    }

    protected void startAsyncOps(NaspServerSocketChannel naspServerSocketChannel, int selectedKeyId, int runningOps, int interestOps) {
        if ((interestOps & SelectionKey.OP_ACCEPT) != 0) {
            naspServerSocketChannel.socket().getNaspTcpListener().startAsyncAccept(ski.hashCode(), selector);
            runningAsyncOps.put(selectedKeyId, runningOps | SelectionKey.OP_ACCEPT);
        }
    }
}

public class NaspSelectorProvider extends SelectorProviderImpl {

    static {
        try {
            boolean sendfileTurnedOff = false;
            loop:
            for (Field field : FileChannelImpl.class.getDeclaredFields()) {
                switch (field.getName()) {
                    case "transferToNotSupported":
                        field.setAccessible(true);
                        field.setBoolean(null, true);
                        field.setAccessible(false);
                        sendfileTurnedOff = true;
                        break loop;
                    case "transferSupported":
                        field.setAccessible(true);
                        field.setBoolean(null, false);
                        field.setAccessible(false);
                        sendfileTurnedOff = true;
                        break loop;
                }
            }
            if (!sendfileTurnedOff) {
                throw new IllegalStateException("couldn't turn off sendfile support");
            }
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    @Override
    public AbstractSelector openSelector() throws IOException {
        return new NaspSelector(this);
    }

    @Override
    public ServerSocketChannel openServerSocketChannel() throws IOException {
        return new NaspServerSocketChannel(this);
    }

    @Override
    public SocketChannel openSocketChannel() throws IOException {
        return new NaspSocketChannel(this);
    }
}
