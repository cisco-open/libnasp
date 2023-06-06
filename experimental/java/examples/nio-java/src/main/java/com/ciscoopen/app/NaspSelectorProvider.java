package com.ciscoopen.app;

import nasp.Nasp;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import sun.nio.ch.FileChannelImpl;
import sun.nio.ch.SelectionKeyImpl;
import sun.nio.ch.SelectorImpl;
import sun.nio.ch.SelectorProviderImpl;

import java.io.IOException;
import java.lang.reflect.Field;
import java.net.InetSocketAddress;
import java.nio.ByteBuffer;
import java.nio.channels.SelectionKey;
import java.nio.channels.Selector;
import java.nio.channels.ServerSocketChannel;
import java.nio.channels.SocketChannel;
import java.nio.channels.spi.AbstractSelector;
import java.nio.channels.spi.SelectorProvider;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.function.Consumer;

class NaspSelector extends SelectorImpl {

    static int getOperation(long selectedKey) {
        return (int) (selectedKey >> 32 & 0xFF);
    }

    static int getSelectedKeyId(long selectedKey) {
        return (int) selectedKey;
    }

    private final nasp.Selector selector = new nasp.Selector();
    private final Map<Integer, SelectionKeyImpl> selectionKeyTable = new HashMap<>();

    protected NaspSelector(SelectorProvider sp) {
        super(sp);
    }

    public nasp.Selector getSelector() {
        return selector;
    }

    @Override
    protected int doSelect(Consumer<SelectionKey> action, long timeout) throws IOException {
        int to = (int) Math.min(timeout, Integer.MAX_VALUE); // max poll timeout
        boolean blocking = (to != 0);

        byte[] selectedKeysNative;
        try {
            begin(blocking);
            selectedKeysNative = nativeSelect(timeout);
        } finally {
            end(blocking);
        }

        if (selectedKeysNative == null) {
            return 0;
        }

        processDeregisterQueue();

        ByteBuffer selectedKeys = ByteBuffer.wrap(selectedKeysNative);

        int numKeysUpdated = 0;
        while (selectedKeys.remaining() > 0) {
            long selectedKey = selectedKeys.getLong();
            int selectedKeyId = getSelectedKeyId(selectedKey);

            SelectionKeyImpl selectionKey = selectionKeyTable.get(selectedKeyId);
            if (selectionKey != null) {
                if (selectionKey.isValid()) {
                    numKeysUpdated += processReadyEvents(getOperation(selectedKey), selectionKey, action);
                }
            }
        }

        return numKeysUpdated;
    }

    private byte[] nativeSelect(long timeout) {
        return selector.select(timeout);
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
        selectionKeyTable.remove(ski.hashCode());
    }

    @Override
    protected void setEventOps(SelectionKeyImpl ski) {
        int interestOps = ski.interestOps();
        if ((interestOps & SelectionKey.OP_WRITE) == 0)
        {
            selector.unregisterWriter(ski.hashCode());
        }
        handleAsyncOps(ski);
    }

    protected void handleAsyncOps(SelectionKeyImpl ski) {

        if (ski.channel() instanceof NaspServerSocketChannel naspServerSockChan) {
            handleAsyncOps(naspServerSockChan, ski.hashCode(), ski.interestOps());
            return;
        }

        if (ski.channel() instanceof NaspSocketChannel naspSockChan) {
            handleAsyncOps(naspSockChan, ski.hashCode(), ski.interestOps());
        }
    }

    protected void handleAsyncOps(NaspSocketChannel naspSocketChannel, int selectedKeyId, int interestedOps) {
        if ((interestedOps & SelectionKey.OP_READ) != 0) {
            naspSocketChannel.getConnection().startAsyncRead(selectedKeyId, selector);
        }

        if ((interestedOps & SelectionKey.OP_WRITE) != 0) {
            naspSocketChannel.getConnection().startAsyncWrite(selectedKeyId, selector);
        }
        
        if ((interestedOps & SelectionKey.OP_CONNECT) != 0) {
            //This could happen if we are servers not clients
            if (naspSocketChannel.getNaspTcpDialer() != null) {
                InetSocketAddress address = naspSocketChannel.getAddress();
                if (address != null) {
                    naspSocketChannel.getNaspTcpDialer().startAsyncDial(selectedKeyId, selector,
                            address.getHostString(), address.getPort());
                } else {
                    naspSocketChannel.setSelector(this);
                }
            }
        }
    }

    protected void handleAsyncOps(NaspServerSocketChannel naspServerSocketChannel, int selectedKeyId, int interestedOps) {
        if ((interestedOps & SelectionKey.OP_ACCEPT) != 0) {
            naspServerSocketChannel.socket().getNaspTcpListener().startAsyncAccept(selectedKeyId, selector);
        }
    }
}

public class NaspSelectorProvider extends SelectorProviderImpl {
    private static final Logger logger = LoggerFactory.getLogger(Nasp.class);
    private static final NaspLogSink naspLogSink = new NaspLogSink(logger);
    private static final ScheduledExecutorService executorService = Executors.newSingleThreadScheduledExecutor();

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

        long naspLogLevel = 0;
        if (logger.isTraceEnabled()) {
            naspLogLevel = 0;
        } else if (logger.isDebugEnabled()) {
            naspLogLevel = 1;
        } else if (logger.isInfoEnabled()) {
            naspLogLevel = 2;
        } else if (logger.isWarnEnabled()) {
            naspLogLevel = 3;
        } else if (logger.isErrorEnabled()) {
            naspLogLevel = 4;
        }

        Nasp.setup(naspLogLevel);

        executorService.scheduleAtFixedRate(() -> {
            byte[] logBatch = Nasp.nextLogBatchJSON(10);
            if (logBatch != null) {
                naspLogSink.Log(logBatch);
            }

        }, 2000, 100, TimeUnit.MILLISECONDS);
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
