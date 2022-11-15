package com.ciscoopen.nasp;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Configuration;

@Configuration
@ConfigurationProperties(prefix = "nasp")
public class NaspConfiguration {

    private String heimdallURL;
    private String heimdallClientID;

    private String heimdallClientSecret;

    public String getHeimdallURL() {
        return heimdallURL;
    }

    public void setHeimdallURL(String heimdallURL) {
        this.heimdallURL = heimdallURL;
    }

    public String getHeimdallClientID() {
        return heimdallClientID;
    }

    public void setHeimdallClientID(String heimdallClientID) {
        this.heimdallClientID = heimdallClientID;
    }

    public String getHeimdallClientSecret() {
        return heimdallClientSecret;
    }

    public void setHeimdallClientSecret(String heimdallClientSecret) {
        this.heimdallClientSecret = heimdallClientSecret;
    }
}
