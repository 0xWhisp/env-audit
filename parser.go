package main

// ParseResult contains parsed entries and any issues found
type ParseResult struct {
	Entries    map[string]string
	Duplicates []string
	Errors     []error
}

// ParseEnvFile reads and parses a .env file
func ParseEnvFile(path string) (*ParseResult, error) {
	// TODO: Implement in task 4
	return &ParseResult{
		Entries:    make(map[string]string),
		Duplicates: []string{},
		Errors:     []error{},
	}, nil
}

// FormatConfig outputs config as KEY=VALUE lines with redaction
func FormatConfig(entries map[string]string) string {
	// TODO: Implement in task 4
	return ""
}
