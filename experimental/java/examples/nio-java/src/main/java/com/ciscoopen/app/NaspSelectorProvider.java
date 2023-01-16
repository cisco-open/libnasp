package com.ciscoopen.app;



import sun.nio.ch.SelectionKeyImpl;
import sun.nio.ch.SelectorImpl;
import sun.nio.ch.SelectorProviderImpl;

import java.io.IOException;
import java.nio.channels.SelectionKey;
import java.nio.channels.Selector;
import java.nio.channels.ServerSocketChannel;
import java.nio.channels.spi.AbstractSelector;
import java.nio.channels.spi.SelectorProvider;
import java.util.function.Consumer;

class NaspSelector extends SelectorImpl {

    private final nasp.Selector selector = new nasp.Selector();

    protected NaspSelector(SelectorProvider sp) {
        super(sp);
    }

    @Override
    protected int doSelect(Consumer<SelectionKey> action, long timeout) throws IOException {
        return (int) selector.select(timeout);
    }

    @Override
    public Selector wakeup() {
        return null;
    }

    @Override
    protected void implRegister(SelectionKeyImpl ski) {
        super.implRegister(ski);
        ski.channel();
    }

    @Override
    protected void implClose() throws IOException {

    }

    @Override
    protected void implDereg(SelectionKeyImpl ski) throws IOException {

    }

    @Override
    protected void setEventOps(SelectionKeyImpl ski) {

    }
}

public class NaspSelectorProvider extends SelectorProviderImpl {

    @Override
    public AbstractSelector openSelector() throws IOException {
        return new NaspSelector(this);
    }

    @Override
    public ServerSocketChannel openServerSocketChannel() throws IOException {
        return new NaspServerSocketChannel(this);
    }
}
