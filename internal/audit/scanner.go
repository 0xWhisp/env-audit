package audit

// Result aggregates all audit findings
type Result struct {
	Issues   []Issue
	HasRisks bool
	Summary  map[IssueType]int
}

// ScanOptions configures the scan behavior
type ScanOptions struct {
	Required   []string
	Ignore     []string
	Duplicates []string
	CheckLeaks bool
	Strict     bool
}

// Scan runs all checks and returns aggregated results
func Scan(env map[string]string, opts *ScanOptions) *Result {
	if opts == nil {
		opts = &ScanOptions{}
	}

	var issues []Issue

	// Run all checks
	issues = append(issues, CheckEmpty(env, opts.Ignore)...)
	issues = append(issues, CheckMissing(env, opts.Required, opts.Ignore)...)
	issues = append(issues, CheckSensitive(env, opts.Ignore)...)

	// Add duplicate issues
	ignoreSet := toSet(opts.Ignore)
	for _, key := range opts.Duplicates {
		if ignoreSet[key] {
			continue
		}
		issues = append(issues, Issue{
			Type:    IssueDuplicate,
			Key:     key,
			Message: "duplicate key definition",
		})
	}

	// Build summary
	summary := make(map[IssueType]int)
	for _, issue := range issues {
		summary[issue.Type]++
	}

	return &Result{
		Issues:   issues,
		HasRisks: len(issues) > 0,
		Summary:  summary,
	}
}
