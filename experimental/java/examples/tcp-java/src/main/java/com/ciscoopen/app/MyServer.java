package com.ciscoopen.app;

import java.io.*;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.net.SocketAddress;
import java.net.UnknownHostException;
import java.util.Scanner;

public class MyServer {
    public static void main(String[] args) {
        initializeServer();
//        initializeClient();
    }
    public static void initializeClient() {
        try (WrappedClientSocket wrappedSocket = new WrappedClientSocket()) {
            wrappedSocket.connect(new InetSocketAddress("localhost", 10000));
            InputStream input = wrappedSocket.getInputStream();
            OutputStream output = wrappedSocket.getOutputStream();

            Scanner scanner  = new Scanner(input);
            PrintWriter writer = new PrintWriter(new OutputStreamWriter(output), true);
            String message = "hello";
            writer.println(message);
            System.out.println("Request: " + message);
            if (scanner.hasNextLine()) {
                String line = scanner.nextLine();
                System.out.println("Received: " + line);
            }
            scanner.close();

        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    public static void initializeServer() {
        try (WrappedServerSocket serverSocket = new WrappedServerSocket(10000)) {
            boolean done = false;
            while(!done) {
                WrappedSocket connectionSocket = (WrappedSocket) serverSocket.accept();
                WrappedInputStream inputToServer = (WrappedInputStream) connectionSocket.getInputStream();
                WrappedOutputStream outputFromServer = (WrappedOutputStream) connectionSocket.getOutputStream();

                Scanner scanner = new Scanner(inputToServer);
                PrintWriter serverPrintOut = new PrintWriter(new OutputStreamWriter(outputFromServer), true);
                if (scanner.hasNextLine()) {
                    String line = scanner.nextLine();
                    serverPrintOut.println(line);
                    if (line.toLowerCase().trim().equals("exit")) {
                        done = true;
                    }
                }
                scanner.close();
            }
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}