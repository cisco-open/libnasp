package com.ciscoopen.nasp;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class HelloController {

    @GetMapping("/")
    public String index(String body) {
        return "Greetings from Spring Boot!";
    }
}
