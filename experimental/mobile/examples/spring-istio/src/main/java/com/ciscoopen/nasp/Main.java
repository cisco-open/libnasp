package com.ciscoopen.nasp;

import istio.HTTPResponse;
import istio.HTTPTransport;
import istio.Istio;

public class Main {
    public static void main(String[] args) throws Exception {
        HTTPTransport transport = Istio.newHTTPTransport("https://localhost:16443/config", "test-http-16362813-F46B-41AC-B191-A390DB1F6BDF", "16362813-F46B-41AC-B191-A390DB1F6BDF");
        HTTPResponse response = transport.request("GET", "https://echo.testing", null);
        System.out.println(new String(response.getBody()));
    }
}
