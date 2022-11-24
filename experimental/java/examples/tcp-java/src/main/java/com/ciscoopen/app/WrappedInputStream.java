package com.ciscoopen.app;

import nasp.Connection;

import java.io.IOException;
import java.io.InputStream;

public class WrappedInputStream extends InputStream {
    Connection conn;

    public WrappedInputStream(Connection conn) {
        this.conn = conn;
    }

    @Override
    public int read() throws IOException {
        byte[] buffer = new byte[1];
        try {
            Long len = this.conn.read(buffer);
            return buffer[0];
        } catch (Exception e) {
            if ("EOF".equals(e.getMessage())) {
                return -1;
            }
            throw new IOException(e);
        }
    }

//    @Override
//    public int read(byte b[]) throws IOException {
//        try {
//            return (int)this.conn.read(b);
//        }catch (Exception e) {
//            throw new IOException("could not read from nasps socket");
//        }
//    }

}
