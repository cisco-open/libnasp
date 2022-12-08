package com.ciscoopen.app;

import java.io.IOException;
import java.io.OutputStreamWriter;
import java.io.PrintWriter;
import java.util.Scanner;

public class ServerThread implements Runnable{
    private WrappedSocket sock;

    ServerThread(WrappedSocket sock){
        this.sock = sock;
    }
    @Override
    public void run() {
        try {
            WrappedInputStream inputToServer = (WrappedInputStream) this.sock.getInputStream();
            WrappedOutputStream outputFromServer = (WrappedOutputStream) this.sock.getOutputStream();
            Scanner scanner = new Scanner(inputToServer);
            PrintWriter serverPrintOut = new PrintWriter(new OutputStreamWriter(outputFromServer), true);
            while (scanner.hasNextLine()) {
                String line = scanner.nextLine();
                serverPrintOut.println(line);
            }
            scanner.close();
        } catch (IOException e) {
            e.printStackTrace();
        }
    }
}
