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

// IsWarning returns true if the issue type is a warning (not an error)
func (t IssueType) IsWarning() bool {
	switch t {
	case IssueEmpty, IssueDuplicate, IssueExtra:
		return true
	default:
		return false
	}
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

	// Determine HasRisks based on strict mode
	hasRisks := hasRiskIssues(issues, opts.Strict)

	return &Result{
		Issues:   issues,
		HasRisks: hasRisks,
		Summary:  summary,
	}
}

// hasRiskIssues returns true if there are issues that should cause exit code 1
// In strict mode, warnings are treated as errors
func hasRiskIssues(issues []Issue, strict bool) bool {
	for _, issue := range issues {
		// Info-level issues (IssueSensitive) never cause risks
		if issue.Type == IssueSensitive {
			continue
		}
		// Errors always cause risks
		if !issue.Type.IsWarning() {
			return true
		}
		// Warnings cause risks only in strict mode
		if strict {
			return true
		}
	}
	return false
}
