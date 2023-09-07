package pii

import (
	"bytes"
	"regexp"
)

var patterns = map[string]string{
	"IP V4 Address":           `\b((25[0-5]|(2[0-4]|1\d|[1-9]?)\d)\.?\b){4}\b`,
	"IP V6 Address":           `\b((([0-9A-Fa-f]{1,4}:){1,6}:)|(([0-9A-Fa-f]{1,4}:){7}))([0-9A-Fa-f]{1,4})\b`,
	"Email":                   `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
	"US Social Security No":   `\b\d{3}-\d{2}-\d{4}\b`,
	"Credit Card Number":      `\b\d{4}-\d{4}-\d{4}-\d{4}\b`,
	"Phone Number":            `\b(\+\d{1,2}\s)?\(?\d{3}\)?[\s.-]\d{3}[\s.-]\d{4}\b`,
	"AWS Access ID":           `\b(AKIA|A3T|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[A-Z0-9]{12,}\b`,
	"MAC Address":             `\b((([a-zA-z0-9]{2}[-:]){5}([a-zA-z0-9]{2}))|(([a-zA-z0-9]{2}:){5}([a-zA-z0-9]{2})))\b`,
	"Google API Key":          `\bAIza[0-9A-Za-z-_]{35}\b`,
	"Google OAuth ID":         `\b[0-9]+-[0-9A-Za-z_]{32}\.apps\.googleusercontent\.com\b`,
	"Github Token":            `\b(ghp|gho|ghu|ghs)_\w{36}\b`,
	"Heroku API Key":          `\b[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\b`,
	"Stripe Key":              `\b(?:r|s)k_(test|live)_[0-9a-zA-Z]{24}\b`,
	"Firebase Auth Domain":    `\b([a-z0-9-]){1,30}(\.firebaseapp\.com)\b`,
	"Slack App Token":         `\bxapp-[0-9]+-[A-Za-z0-9_]+-[0-9]+-[a-f0-9]+\b`,
	"Github Fine Grained PAT": `\bgithub_pat_[A-Za-z0-9_]{82}\b`,
	"AWS Env Var":             `AWS_SECRET_ACCESS_KEY=([A-Za-z0-9/+=]+)`,
	"GCP Env Var":             `GOOGLE_APPLICATION_CREDENTIALS=([A-Za-z0-9/+=]+)`,
	"Azure Env Var":           `AZURE_CLIENT_SECRET=([A-Za-z0-9/+=]+)`,
	"Database Password":       `DATABASE_URL=(?:postgres|mysql)://\w+:(\w+)@\w+`,
}

// DetectPII searches `text` for common PII patterns and obfuscates them depending on the `obfuscate` parameter
// the return boolean indicates whether a PII was found or not
func DetectPII(text []byte, obfuscate bool) bool {
	piiDetected := false

	for secretType, pattern := range patterns {
		if detect(secretType, pattern, text, obfuscate) {
			piiDetected = true
		}
	}

	return piiDetected
}

func detect(secretType, pattern string, text []byte, obfuscate bool) bool {
	r := regexp.MustCompile(pattern)
	matches := r.FindAllIndex(text, -1)

	if len(matches) > 0 {
		if obfuscate {
			for _, match := range matches {
				start, end := match[0], match[1]

				// Check if it's an env var e.g. "KEY=value"
				isEnvVar := bytes.Contains(text[start:end], []byte("="))

				if isEnvVar {
					// For env vars, only obfuscate after the "=", for everything else, obfuscate the entire match
					equalsPos := bytes.Index(text[start:end], []byte("=")) + start
					start = equalsPos + 1
				}

				for i := start; i < end; i++ {
					text[i] = '*'
				}
			}
		}
		return true
	}

	return false
}
