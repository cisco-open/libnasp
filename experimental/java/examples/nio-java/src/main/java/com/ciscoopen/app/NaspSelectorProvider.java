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
        if (ski.channel() instanceof NaspServerSocketChannel naspServerSockChan) {
            naspServerSockChan.socket().getTCPListener().startAsyncAccept(ski.hashCode(), selector);
        } else if (ski.channel() instanceof NaspSocketChannel naspSockChan) {
            int interestOps = ski.interestOps();
            if ((interestOps & SelectionKey.OP_READ) != 0) {
                naspSockChan.getConnection().startAsyncRead(ski.hashCode(), selector);
            }
            if ((interestOps & SelectionKey.OP_WRITE) != 0) {
                naspSockChan.getConnection().startAsyncWrite(ski.hashCode(), selector);
            }
            if ((interestOps & SelectionKey.OP_CONNECT) != 0) {
                //This could happen if we are servers not clients
                if (naspSockChan.getTCPDialer() != null) {
                    InetSocketAddress address = naspSockChan.getAddress();
                    naspSockChan.getTCPDialer().startAsyncDial(ski.hashCode(), selector,
                            address.getHostString(), address.getPort());
                }
            }
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
