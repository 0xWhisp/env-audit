package main

// ScanResult aggregates all audit findings
type ScanResult struct {
	Issues   []Issue
	HasRisks bool
}

// Scan runs all checks and returns aggregated results
func Scan(env map[string]string, required []string, duplicates []string) *ScanResult {
	// TODO: Implement in task 6
	return &ScanResult{}
}
