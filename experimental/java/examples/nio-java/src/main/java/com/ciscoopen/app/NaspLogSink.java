package com.ciscoopen.app;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import nasp.Nasp;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.slf4j.MDC;

import java.io.IOException;
import java.util.Iterator;
import java.util.Map;

public class NaspLogSink {
    protected final ObjectMapper objectMapper = new ObjectMapper();
    protected final Logger logger;

    public NaspLogSink(Logger logger) {
        this.logger = logger;
    }

    public void Log(byte[] logBatchJSON) {
        try {
            JsonNode jsonNode = objectMapper.readTree(logBatchJSON);

            Iterator<JsonNode> it = jsonNode.elements();
            while (it.hasNext()) {
                StringBuilder sb = new StringBuilder();
                JsonNode logLine = it.next();

                if (logLine.has("message")) {
                    sb.append(logLine.get("message").asText());
                }

                Iterator<Map.Entry<String, JsonNode>> logLineFields = logLine.fields();
                while (logLineFields.hasNext()) {
                    Map.Entry<String, JsonNode> field = logLineFields.next();
                    if (field.getKey() == "level" || field.getKey() == "message") {
                        continue;
                    }

                    if (sb.length() > 0) {
                        sb.append(" ");
                    }
                    sb.append(field.getKey()).append("=").append(field.getValue().asText());
                }
                String logMessage = sb.toString();

                String level = "";
                if (logLine.has("level")) {
                    level = logLine.get("level").asText();
                }

                switch (level) {
                    case "panic":
                    case "fatal":
                    case "error":
                        logger.error(logMessage);
                        break;
                    case "warn":
                        logger.warn(logMessage);
                    case "info":
                        logger.info(logMessage);
                        break;
                    case "debug":
                        logger.debug(logMessage);
                        break;
                    case "trace":
                        logger.trace(logMessage);
                        break;
                    default: // ignore
                }
            }

        } catch (IOException e) {
            e.printStackTrace();
        }
    }
}
