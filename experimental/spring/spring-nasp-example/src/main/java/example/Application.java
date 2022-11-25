package example;

import com.cisco.nasp.spring.NaspConfiguration;
import com.cisco.nasp.spring.NaspWebServerFactory;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.web.reactive.server.ReactiveWebServerFactory;
import org.springframework.context.annotation.Bean;

@SpringBootApplication(scanBasePackageClasses = {NaspConfiguration.class})
public class Application {
    @Bean
    public ReactiveWebServerFactory webServerFactory(NaspConfiguration configuration) {
        return new NaspWebServerFactory(configuration);
    }

    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
    }
}
