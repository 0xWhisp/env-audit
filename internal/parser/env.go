package parser

import (
	"bufio"
	"os"
	"strings"

	"env-audit/internal/audit"
)

// ParseResult contains parsed entries and any issues found
type ParseResult struct {
	Entries    map[string]string
	Duplicates []string
	Errors     []error
}

// ParseEnvFile reads and parses a .env file
func ParseEnvFile(path string) (*ParseResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := &ParseResult{
		Entries:    make(map[string]string),
		Duplicates: []string{},
		Errors:     []error{},
	}

	seen := make(map[string]bool)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Find the first = sign
		idx := strings.Index(line, "=")
		if idx == -1 {
			continue // Skip malformed lines
		}

		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		// Handle quoted values
		value = unquote(value)

		// Track duplicates
		if seen[key] {
			result.Duplicates = append(result.Duplicates, key)
		}
		seen[key] = true

		result.Entries[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}


// unquote removes surrounding quotes from a value
func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// FormatEnv outputs config as KEY=VALUE lines with optional redaction
func FormatEnv(entries map[string]string, redact bool) string {
	var lines []string
	for key, value := range entries {
		if redact && audit.IsSensitiveKey(key) {
			lines = append(lines, key+"=[REDACTED]")
		} else {
			lines = append(lines, key+"="+value)
		}
	}
	return strings.Join(lines, "\n")
}
