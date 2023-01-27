package com.ciscoopen.app;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.nio.ByteBuffer;
import java.nio.channels.SelectionKey;
import java.nio.channels.Selector;
import java.nio.channels.ServerSocketChannel;
import java.nio.channels.SocketChannel;
import java.util.Iterator;
import java.util.Set;

public class MyServer {
    public static void main(String[] args) {
//        initializeServer();
        initializeClient();
    }

    public static void initializeClient() {
        try {
            Selector selector = Selector.open();
            SocketChannel client = SocketChannel.open();
            client.configureBlocking(false);
            client.connect(new InetSocketAddress("localhost", 10000));
            client.register(selector, SelectionKey.OP_CONNECT);
            System.out.println("Connection Set: " + client.getRemoteAddress());

            ByteBuffer buffer = ByteBuffer.allocate(256);
            while (true) {
                int channelCount = selector.select();
                if (channelCount > 0 ){
                    Set<SelectionKey> keys = selector.selectedKeys();
                    Iterator<SelectionKey> iterator = keys.iterator();
                    while (iterator.hasNext()) {
                        SelectionKey key = iterator.next();
                        iterator.remove();
                        if (key.isConnectable()) {
                            client.finishConnect();
                            key.interestOps(SelectionKey.OP_WRITE);
                        } else if (key.isWritable()){
                            String newData = "This is string from the client!";
                            buffer.put(newData.getBytes());
                            buffer.flip();
                            client.write(buffer);
                            buffer.clear();
                            key.interestOps(SelectionKey.OP_READ);
                        } else if (key.isReadable()){
                            client.read(buffer);
                            System.out.println(new String(buffer.array()));
                            key.cancel();
                            client.close();
                        }
                    }
                }
            }

        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    public static void initializeServer() {
        try {
            ServerSocketChannel server = ServerSocketChannel.open();

            server.socket().bind(new InetSocketAddress("localhost", 10000));
            server.socket().setReuseAddress(true);
            server.configureBlocking(false);

            Selector selector = Selector.open();
            server.register(selector, server.validOps());

            ByteBuffer buffer = ByteBuffer.allocate(4096);
            while (true) {
                int channelCount = selector.select();
                if (channelCount > 0) {
                    Set<SelectionKey> keys = selector.selectedKeys();
                    Iterator<SelectionKey> iterator = keys.iterator();
                    while (iterator.hasNext()) {
                        SelectionKey key = iterator.next();
                        iterator.remove();
                        if (key.isAcceptable()) {
                            SocketChannel client = server.accept();
                            client.configureBlocking(false);
                            client.register(selector, client.validOps(), client.socket().getPort());
                        } else if (key.isReadable()) {
                            SocketChannel client = (SocketChannel) key.channel();
                            if (client.read(buffer) < 0) {
                                key.cancel();
                                client.close();
                            } else {
                                buffer.flip(); // read from the buffer
                                client.write(buffer);
                                buffer.clear();
                            }
                        }
                    }
                }
            }
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    private static void p(String s) {
        System.out.println(s);
    }
}