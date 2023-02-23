package com.cisco.nasp.spring;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Configuration;

@Configuration
@ConfigurationProperties(prefix = "nasp")
public class NaspConfiguration {

    private String heimdallURL;
    private String heimdallAuthorizationToken;

    public String getHeimdallURL() {
        return heimdallURL;
    }

    public void setHeimdallURL(String heimdallURL) {
        this.heimdallURL = heimdallURL;
    }

    public String getHeimdallAuthorizationToken() {
        return heimdallAuthorizationToken;
    }

    public void setHeimdallAuthorizationToken(String heimdallAuthorizationToken) {
        this.heimdallAuthorizationToken = heimdallAuthorizationToken;
    }
}
