package com.cisco.nasp.spring.example;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

import java.util.concurrent.atomic.AtomicInteger;

@RestController
public class HelloController {

    private final AtomicInteger counter = new AtomicInteger();

    @GetMapping("/")
    public String index() {
        return "Greetings from Spring Boot!";
    }

    @PostMapping("/echo")
    public String echo(@RequestBody String body) {
        return body;
    }


    @GetMapping("/random")
    public String random() {
        if (counter.incrementAndGet() % 10 == 0) {
            throw new IllegalStateException("Every 10th request is an error my friend!");
        }
        return "You visited me " + counter.get() + " time(s) today";
    }
}
