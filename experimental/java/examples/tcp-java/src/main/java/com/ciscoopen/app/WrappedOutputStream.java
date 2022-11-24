package com.ciscoopen.app;

import nasp.Connection;

import java.io.IOException;
import java.io.OutputStream;
import java.nio.ByteBuffer;

public class WrappedOutputStream extends OutputStream {
    Connection conn;

    public WrappedOutputStream(Connection conn) {
        this.conn = conn;
    }
    
    @Override
    public void write(int b) throws IOException {
        final byte[] buffer = new byte[]{(byte)b};
        try {
            this.conn.write(buffer);
        } catch (Exception e) {
            throw new IOException("could not write to nasps socket");
        }
    }

    @Override
    public void write(byte b[]) throws IOException {
        try {
            this.conn.write(b);
        }catch (Exception e) {
            throw new IOException("could not write to nasps socket");
        }
    }
    
}
