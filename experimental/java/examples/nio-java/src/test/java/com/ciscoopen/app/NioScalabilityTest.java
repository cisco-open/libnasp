package com.ciscoopen.app;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.ServerSocket;
import java.net.StandardSocketOptions;
import java.nio.ByteBuffer;
import java.nio.channels.CancelledKeyException;
import java.nio.channels.ReadableByteChannel;
import java.nio.channels.SelectionKey;
import java.nio.channels.Selector;
import java.nio.channels.ServerSocketChannel;
import java.nio.channels.SocketChannel;
import java.nio.channels.WritableByteChannel;
import java.nio.channels.spi.SelectorProvider;
import java.util.Iterator;
import java.util.Random;
import java.util.concurrent.BlockingDeque;
import java.util.concurrent.Executor;
import java.util.concurrent.Executors;
import java.util.concurrent.LinkedBlockingDeque;

/**
 * Modified and taken from https://gist.github.com/jbrisbin/1557225
 */
public class NioScalabilityTest {

    private static final boolean USE_NASP = true;

    public static void main(String[] args) throws InterruptedException, IOException {
        final int cores = Runtime.getRuntime().availableProcessors();
        final int bufferSize = 4 * 1024;
        final int ringSize = Double.valueOf(Math.pow(2, cores + 1)).intValue();
        final Random random = new Random();

        final BlockingDeque<SelectionEvent> readyEvents = new LinkedBlockingDeque<>(ringSize);
        for (int i = 0; i < ringSize; i++) {
            readyEvents.add(new SelectionEvent());
        }

        TaskRunner[] runners = new TaskRunner[cores];
        for (int i = 0; i < cores; i++) {
            runners[i] = new TaskRunner(ringSize, readyEvents);
        }

        SelectorProvider selectorProvider;
        if (USE_NASP)
            selectorProvider = new NaspSelectorProvider();
        else
            selectorProvider = new sun.nio.ch.PollSelectorProvider();

        ServerSocketChannel serverSocketChannel = selectorProvider.openServerSocketChannel();

        serverSocketChannel.configureBlocking(false);

        Selector selector = selectorProvider.openSelector();
        serverSocketChannel.socket().bind(new InetSocketAddress("127.0.0.1", 3000), 1024);
        System.out.println("Listening on localhost:3000...");
        serverSocketChannel.register(selector, SelectionKey.OP_ACCEPT);

        while (true) {
            int cnt = selector.select();
            if (cnt > 0) {
                Iterator<SelectionKey> keys = selector.selectedKeys().iterator();
                while (keys.hasNext()) {
                    SelectionKey key = keys.next();
                    keys.remove();
                    if (key.isValid()) {
                        if (key.isAcceptable()) {
                            ServerSocket serverSocket = serverSocketChannel.socket();
                            serverSocket.setReceiveBufferSize(bufferSize);
                            serverSocket.setReuseAddress(true);

                            boolean hasSocket = true;
                            do {
                                SocketChannel channel = serverSocketChannel.accept();
                                if (null != channel) {
                                    channel.configureBlocking(false);
                                    channel.setOption(StandardSocketOptions.SO_KEEPALIVE, true);
                                    channel.setOption(StandardSocketOptions.TCP_NODELAY, true);
                                    channel.setOption(StandardSocketOptions.SO_RCVBUF, bufferSize);
                                    channel.setOption(StandardSocketOptions.SO_SNDBUF, bufferSize);
                                    SelectionKey readKey = channel.register(selector, SelectionKey.OP_READ | SelectionKey.OP_WRITE);

                                    TaskRunner runner = runners[random.nextInt(cores)];
                                    final SelectionEvent event = readyEvents.take();
                                    event.channel = channel;
                                    event.selector = selector;
                                    event.key = readKey;
                                    event.serverChannel = serverSocketChannel;

                                    runner.events.add(event);
                                } else {
                                    hasSocket = false;
                                }
                            } while (hasSocket);
                        } else if (key.isReadable()) {
                            TaskRunner runner = runners[random.nextInt(cores)];
                            final SelectionEvent event = readyEvents.take();
                            event.channel = (SocketChannel) key.channel();
                            event.selector = selector;
                            event.key = key;
                            event.serverChannel = serverSocketChannel;

                            runner.events.add(event);
                        }
                    }
                }
            }
        }
    }

    static int safeRead(ReadableByteChannel channel, ByteBuffer dst) throws IOException {
        int read = -1;
        try {
            // Read data from the Channel
            read = channel.read(dst);
        } catch (IOException e) {
            switch ("" + e.getMessage()) {
                case "null":
                case "Connection reset by peer":
                case "Broken pipe":
                case "Connection reset":
                    break;
                default:
                    e.printStackTrace();
            }
            channel.close();
        } catch (CancelledKeyException e) {
            channel.close();
        }
        return read;
    }

    static int safeWrite(WritableByteChannel channel, ByteBuffer src) throws IOException {
        int written = -1;
        try {
            // Write the response immediately
            written = channel.write(src);
        } catch (IOException e) {
            switch ("" + e.getMessage()) {
                case "null":
                case "Connection reset by peer":
                case "Broken pipe":
                case "Connection reset":
                    break;
                default:
                    e.printStackTrace();
            }
            channel.close();
        } catch (CancelledKeyException e) {
            channel.close();
        }
        return written;
    }

    static class SelectionEvent {

        Selector selector;
        ServerSocketChannel serverChannel;
        SelectionKey key;
        SocketChannel channel;
        ByteBuffer buffer = ByteBuffer.allocate(16 * 1024);

        SelectionEvent() {
        }

    }

    static class TaskRunner implements Runnable {

        BlockingDeque<SelectionEvent> events;
        BlockingDeque<SelectionEvent> readyEvents;
        Executor executor = Executors.newSingleThreadExecutor();
        ByteBuffer msg = ByteBuffer.wrap(("HTTP/1.1 200 OK\r\n" + "Connection: Keep-Alive\r\n" + "Content-Type: text/plain\r\n" + "Content-Length: 12\r\n\r\n" + "Hello World!").getBytes());

        TaskRunner(int size, BlockingDeque<SelectionEvent> readyEvents) {
            this.events = new LinkedBlockingDeque<>(size);
            this.readyEvents = readyEvents;
            executor.execute(this);
        }

        @Override
        public void run() {
            SelectionEvent ev;
            try {
                while (null != (ev = events.take())) {
                    if (ev.buffer.position() > 0) {
                        ev.buffer.clear();
                    }

                    try {
                        int read = safeRead(ev.channel, ev.buffer);
                        while (read > 0) {
                            safeWrite(ev.channel, msg.duplicate());

                            // Read the data into memory
                            ev.buffer.flip();
                            byte[] bytes = new byte[ev.buffer.remaining()];
                            ev.buffer.get(bytes);
                            //String input = new String(bytes);
                            ev.buffer.clear();
                            read = safeRead(ev.channel, ev.buffer);
                        }
                        if (read < 0) {
                            ev.key.cancel();
                        }
                    } catch (IOException e) {
                        throw new IllegalStateException(e);
                    } finally {
                        readyEvents.push(ev);
                    }
                }
            } catch (InterruptedException e) {
                throw new IllegalStateException(e);
            }
        }
    }
}
