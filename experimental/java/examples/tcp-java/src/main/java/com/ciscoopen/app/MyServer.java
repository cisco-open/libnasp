package com.ciscoopen.app;

import java.io.*;
import java.util.Scanner;

public class MyServer {
    public static void main(String[] args) {
        initializeServer();
    }

    public static void initializeServer() {
        try (WrappedServerSocket serverSocket = new WrappedServerSocket(9090)) {
            WrappedSocket connectionSocket = (WrappedSocket) serverSocket.accept();
            WrappedInputStream inputToServer = (WrappedInputStream) connectionSocket.getInputStream();
            WrappedOutputStream outputFromServer = (WrappedOutputStream) connectionSocket.getOutputStream();
            
            Scanner scanner = new Scanner(inputToServer, "UTF-8");
            PrintWriter serverPrintOut = new PrintWriter(new OutputStreamWriter(outputFromServer, "UTF-8"), true);

            boolean done = false;

            while (!done && scanner.hasNextLine()) {
                String line = scanner.nextLine();
                serverPrintOut.println(line);
                if (line.toLowerCase().trim().equals("exit")) {
                    done = true;
                }
            }
            scanner.close();
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}