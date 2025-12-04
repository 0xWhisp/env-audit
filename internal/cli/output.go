package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"env-audit/internal/audit"
)

// Formatter defines the interface for output formatting
type Formatter interface {
	Format(result *audit.Result) string
}

// JSONFormatter outputs results as JSON
type JSONFormatter struct{}

// GitHubFormatter outputs results in GitHub Actions workflow command format
type GitHubFormatter struct{}

// TextFormatter outputs results with optional color support
type TextFormatter struct {
	UseColor bool
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

// jsonIssue represents an issue in JSON output
type jsonIssue struct {
	Type    string `json:"type"`
	Key     string `json:"key"`
	Message string `json:"message"`
}

// jsonOutput represents the complete JSON output structure
type jsonOutput struct {
	HasRisks bool           `json:"hasRisks"`
	Issues   []jsonIssue    `json:"issues"`
	Summary  map[string]int `json:"summary"`
}

// issueTypeToString converts IssueType to string for JSON
func issueTypeToString(t audit.IssueType) string {
	switch t {
	case audit.IssueEmpty:
		return "empty"
	case audit.IssueMissing:
		return "missing"
	case audit.IssueSensitive:
		return "sensitive"
	case audit.IssueDuplicate:
		return "duplicate"
	case audit.IssueLeak:
		return "leak"
	case audit.IssueExtra:
		return "extra"
	default:
		return "unknown"
	}
}

// Format implements Formatter interface for TextFormatter
// Uses colors for errors (red), warnings (yellow), and success (green)
func (f *TextFormatter) Format(result *audit.Result) string {
	if result == nil || len(result.Issues) == 0 {
		msg := "env-audit scan results\n======================\n\nNo issues found."
		if f.UseColor {
			return colorGreen + msg + colorReset
		}
		return msg
	}

	// Group issues by type
	groups := make(map[audit.IssueType][]audit.Issue)
	for _, issue := range result.Issues {
		groups[issue.Type] = append(groups[issue.Type], issue)
	}

	var sb strings.Builder
	sb.WriteString("env-audit scan results\n")
	sb.WriteString("======================\n")

	// Output each group in order
	typeOrder := []audit.IssueType{audit.IssueEmpty, audit.IssueMissing, audit.IssueSensitive, audit.IssueDuplicate, audit.IssueExtra, audit.IssueLeak}
	typeNames := map[audit.IssueType]string{
		audit.IssueEmpty:     "Empty Values",
		audit.IssueMissing:   "Missing Required",
		audit.IssueSensitive: "Sensitive Keys Detected",
		audit.IssueDuplicate: "Duplicate Keys",
		audit.IssueExtra:     "Extra Variables",
		audit.IssueLeak:      "Potential Leaks",
	}

	for _, t := range typeOrder {
		issues := groups[t]
		if len(issues) == 0 {
			continue
		}

		// Determine color based on issue type
		color := ""
		if f.UseColor {
			if t == audit.IssueMissing || t == audit.IssueLeak {
				color = colorRed
			} else {
				color = colorYellow
			}
		}

		if color != "" {
			sb.WriteString(color)
		}
		sb.WriteString(fmt.Sprintf("\n%s (%d):\n", typeNames[t], len(issues)))
		for _, issue := range issues {
			if t == audit.IssueSensitive {
				sb.WriteString(fmt.Sprintf("  - %s: [REDACTED]\n", issue.Key))
			} else if t == audit.IssueLeak {
				sb.WriteString(fmt.Sprintf("  - %s: %s\n", issue.Key, issue.Message))
			} else {
				sb.WriteString(fmt.Sprintf("  - %s\n", issue.Key))
			}
		}
		if color != "" {
			sb.WriteString(colorReset)
		}
	}

	sb.WriteString(fmt.Sprintf("\nSummary: %d issues found\n", len(result.Issues)))
	return sb.String()
}

// Format implements Formatter interface for GitHubFormatter
// Uses ::error:: for critical issues (missing, leak, duplicate)
// Uses ::warning:: for non-critical issues (empty, sensitive, extra)
func (f *GitHubFormatter) Format(result *audit.Result) string {
	if result == nil || len(result.Issues) == 0 {
		return ""
	}

	var lines []string
	for _, issue := range result.Issues {
		prefix := "::warning::"
		// Critical issues get error level
		if issue.Type == audit.IssueMissing || issue.Type == audit.IssueLeak || issue.Type == audit.IssueDuplicate {
			prefix = "::error::"
		}
		lines = append(lines, fmt.Sprintf("%s%s: %s", prefix, issue.Key, issue.Message))
	}
	return strings.Join(lines, "\n")
}

// Format implements Formatter interface for JSONFormatter
func (f *JSONFormatter) Format(result *audit.Result) string {
	output := jsonOutput{
		HasRisks: false,
		Issues:   []jsonIssue{},
		Summary:  make(map[string]int),
	}

	if result != nil {
		output.HasRisks = result.HasRisks

		for _, issue := range result.Issues {
			output.Issues = append(output.Issues, jsonIssue{
				Type:    issueTypeToString(issue.Type),
				Key:     issue.Key,
				Message: issue.Message,
			})
		}

		for issueType, count := range result.Summary {
			output.Summary[issueTypeToString(issueType)] = count
		}
	}

	data, err := json.Marshal(output)
	if err != nil {
		return `{"hasRisks":false,"issues":[],"summary":{}}`
	}
	return string(data)
}

// FormatSummary produces human-readable output grouped by issue type
func FormatSummary(result *audit.Result) string {
	if result == nil || len(result.Issues) == 0 {
		return "env-audit scan results\n======================\n\nNo issues found.\n"
	}

	// Group issues by type
	groups := make(map[audit.IssueType][]audit.Issue)
	for _, issue := range result.Issues {
		groups[issue.Type] = append(groups[issue.Type], issue)
	}

	var sb strings.Builder
	sb.WriteString("env-audit scan results\n")
	sb.WriteString("======================\n")

	// Output each group in order
	typeOrder := []audit.IssueType{audit.IssueEmpty, audit.IssueMissing, audit.IssueSensitive, audit.IssueDuplicate, audit.IssueExtra, audit.IssueLeak}
	typeNames := map[audit.IssueType]string{
		audit.IssueEmpty:     "Empty Values",
		audit.IssueMissing:   "Missing Required",
		audit.IssueSensitive: "Sensitive Keys Detected",
		audit.IssueDuplicate: "Duplicate Keys",
		audit.IssueExtra:     "Extra Variables",
		audit.IssueLeak:      "Potential Leaks",
	}

	for _, t := range typeOrder {
		issues := groups[t]
		if len(issues) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("\n%s (%d):\n", typeNames[t], len(issues)))
		for _, issue := range issues {
			if t == audit.IssueSensitive {
				sb.WriteString(fmt.Sprintf("  - %s: [REDACTED]\n", issue.Key))
			} else if t == audit.IssueLeak {
				sb.WriteString(fmt.Sprintf("  - %s: %s\n", issue.Key, issue.Message))
			} else {
				sb.WriteString(fmt.Sprintf("  - %s\n", issue.Key))
			}
		}
	}

	sb.WriteString(fmt.Sprintf("\nSummary: %d issues found\n", len(result.Issues)))
	return sb.String()
}

// PrintUsage outputs help text
func PrintUsage(w io.Writer) {
	fmt.Fprintln(w, "env-audit [options]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Options:")
	fmt.Fprintln(w, "  --file, -f <path>     Path to .env file to scan")
	fmt.Fprintln(w, "  --required, -r <vars> Comma-separated list of required variables")
	fmt.Fprintln(w, "  --example, -e <path>  Path to .env.example file for comparison")
	fmt.Fprintln(w, "  --ignore, -i <keys>   Comma-separated list of keys to ignore")
	fmt.Fprintln(w, "  --diff <path>         Compare with another env file")
	fmt.Fprintln(w, "  --dump, -d            Output parsed configuration (with redaction)")
	fmt.Fprintln(w, "  --init                Generate .env.example from current env")
	fmt.Fprintln(w, "  --force               Overwrite existing files")
	fmt.Fprintln(w, "  --json                Output results as JSON")
	fmt.Fprintln(w, "  --github              Output results in GitHub Actions format")
	fmt.Fprintln(w, "  --quiet, -q           Suppress stdout output")
	fmt.Fprintln(w, "  --strict              Treat warnings as errors")
	fmt.Fprintln(w, "  --check-leaks         Analyze values for secret patterns")
	fmt.Fprintln(w, "  --no-color            Disable colored output")
	fmt.Fprintln(w, "  --watch, -w           Watch file for changes")
	fmt.Fprintln(w, "  --version, -V         Show version")
	fmt.Fprintln(w, "  --help, -h            Show this help message")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Exit Codes:")
	fmt.Fprintln(w, "  0  No risks found")
	fmt.Fprintln(w, "  1  Risks detected")
	fmt.Fprintln(w, "  2  Fatal error (invalid arguments, file not found)")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Config File:")
	fmt.Fprintln(w, "  Create .env-audit.yaml or .env-audit.yml in your project root")
	fmt.Fprintln(w, "  CLI flags take precedence over config file values")
}

// Redact returns "[REDACTED]" placeholder
func Redact(value string) string {
	return "[REDACTED]"
}

// ShouldUseColor determines if colored output should be used
// Returns false if:
// - noColor flag is true (--no-color)
// - NO_COLOR env var is set (any value)
// - stdout is not a TTY
func ShouldUseColor(noColor bool, isTTY bool) bool {
	if noColor {
		return false
	}
	// Check NO_COLOR environment variable (https://no-color.org/)
	if _, exists := os.LookupEnv("NO_COLOR"); exists {
		return false
	}
	return isTTY
}
