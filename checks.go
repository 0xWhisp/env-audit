package main

// IssueType represents the category of an audit issue
type IssueType int

const (
	IssueEmpty IssueType = iota
	IssueMissing
	IssueSensitive
	IssueDuplicate
)

// Issue represents a single audit finding
type Issue struct {
	Type    IssueType
	Key     string
	Message string
}

// CheckEmpty finds variables with empty values
func CheckEmpty(env map[string]string) []Issue {
	// TODO: Implement in task 3
	return nil
}

// CheckMissing finds required variables not present
func CheckMissing(env map[string]string, required []string) []Issue {
	// TODO: Implement in task 3
	return nil
}

// CheckSensitive finds keys matching sensitive patterns
func CheckSensitive(env map[string]string) []Issue {
	// TODO: Implement in task 3
	return nil
}

// IsSensitiveKey returns true if key matches sensitive patterns
func IsSensitiveKey(key string) bool {
	// TODO: Implement in task 3
	return false
}
