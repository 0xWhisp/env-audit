package main

// ScanResult aggregates all audit findings
type ScanResult struct {
	Issues   []Issue
	HasRisks bool
}

// Scan runs all checks and returns aggregated results
func Scan(env map[string]string, required []string, duplicates []string) *ScanResult {
	var issues []Issue

	// Run all checks
	issues = append(issues, CheckEmpty(env)...)
	issues = append(issues, CheckMissing(env, required)...)
	issues = append(issues, CheckSensitive(env)...)

	// Add duplicate issues
	for _, key := range duplicates {
		issues = append(issues, Issue{
			Type:    IssueDuplicate,
			Key:     key,
			Message: "duplicate key definition",
		})
	}

	return &ScanResult{
		Issues:   issues,
		HasRisks: len(issues) > 0,
	}
}
