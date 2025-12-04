package cli

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"env-audit/internal/audit"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: env-audit, Property 9: Summary includes all issues**
// **Validates: Requirements 1.3, 4.1**
// For any list of Issues, FormatSummary output SHALL contain the Key of every issue in the list.
func TestProperty_SummaryIncludesAllIssues(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for issue type
	genIssueType := gen.IntRange(0, 3).Map(func(i int) audit.IssueType {
		return audit.IssueType(i)
	})

	// Generator for a single issue with alphanumeric key
	genIssue := gen.Struct(reflect.TypeOf(audit.Issue{}), map[string]gopter.Gen{
		"Type":    genIssueType,
		"Key":     gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		"Message": gen.AnyString(),
	})

	// Generator for slice of issues
	genIssues := gen.SliceOf(genIssue)

	properties.Property("summary contains all issue keys", prop.ForAll(
		func(issues []audit.Issue) bool {
			result := &audit.Result{
				Issues:   issues,
				HasRisks: len(issues) > 0,
			}

			summary := FormatSummary(result)

			// Every issue key must appear in the summary
			for _, issue := range issues {
				if !strings.Contains(summary, issue.Key) {
					return false
				}
			}

			return true
		},
		genIssues,
	))

	properties.TestingRun(t)
}

// **Feature: env-audit-v2, Property 1: JSON output validity and completeness**
// **Validates: Requirements 2.1, 2.2, 2.4**
// For any audit result, when --json flag is used, the output SHALL be valid JSON
// containing all issue keys from the result.
func TestProperty_JSONOutputValidityAndCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for issue type (0-5 covers all IssueType values)
	genIssueType := gen.IntRange(0, 5).Map(func(i int) audit.IssueType {
		return audit.IssueType(i)
	})

	// Generator for a single issue with non-empty alphanumeric key
	genIssue := gen.Struct(reflect.TypeOf(audit.Issue{}), map[string]gopter.Gen{
		"Type":    genIssueType,
		"Key":     gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		"Message": gen.AnyString(),
	})

	// Generator for slice of issues
	genIssues := gen.SliceOf(genIssue)

	properties.Property("JSON output is valid and contains all issue keys", prop.ForAll(
		func(issues []audit.Issue) bool {
			// Build summary from issues
			summary := make(map[audit.IssueType]int)
			for _, issue := range issues {
				summary[issue.Type]++
			}

			result := &audit.Result{
				Issues:   issues,
				HasRisks: len(issues) > 0,
				Summary:  summary,
			}

			formatter := &JSONFormatter{}
			output := formatter.Format(result)

			// Property 1: Output must be valid JSON
			var parsed jsonOutput
			if err := json.Unmarshal([]byte(output), &parsed); err != nil {
				t.Logf("Invalid JSON: %v", err)
				return false
			}

			// Property 2: JSON must contain all issue keys
			if len(parsed.Issues) != len(issues) {
				t.Logf("Issue count mismatch: expected %d, got %d", len(issues), len(parsed.Issues))
				return false
			}

			// Verify each issue key appears in output
			for i, issue := range issues {
				if parsed.Issues[i].Key != issue.Key {
					t.Logf("Key mismatch at index %d: expected %s, got %s", i, issue.Key, parsed.Issues[i].Key)
					return false
				}
			}

			// Property 3: HasRisks must match
			if parsed.HasRisks != result.HasRisks {
				t.Logf("HasRisks mismatch: expected %v, got %v", result.HasRisks, parsed.HasRisks)
				return false
			}

			return true
		},
		genIssues,
	))

	properties.TestingRun(t)
}

func TestRedact(t *testing.T) {
	result := Redact("secret_value")
	if result != "[REDACTED]" {
		t.Errorf("expected [REDACTED], got %s", result)
	}
}

func TestFormatSummary_NilResult(t *testing.T) {
	result := FormatSummary(nil)
	if !strings.Contains(result, "No issues found") {
		t.Error("nil result should show no issues")
	}
}

func TestFormatSummary_EmptyIssues(t *testing.T) {
	result := FormatSummary(&audit.Result{Issues: []audit.Issue{}})
	if !strings.Contains(result, "No issues found") {
		t.Error("empty issues should show no issues")
	}
}

func TestJSONFormatter_NilResult(t *testing.T) {
	f := &JSONFormatter{}
	result := f.Format(nil)
	expected := `{"hasRisks":false,"issues":[],"summary":{}}`
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestJSONFormatter_EmptyIssues(t *testing.T) {
	f := &JSONFormatter{}
	result := f.Format(&audit.Result{
		Issues:   []audit.Issue{},
		HasRisks: false,
		Summary:  map[audit.IssueType]int{},
	})
	expected := `{"hasRisks":false,"issues":[],"summary":{}}`
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestJSONFormatter_WithIssues(t *testing.T) {
	f := &JSONFormatter{}
	result := f.Format(&audit.Result{
		Issues: []audit.Issue{
			{Type: audit.IssueEmpty, Key: "DATABASE_URL", Message: "variable has empty value"},
		},
		HasRisks: true,
		Summary:  map[audit.IssueType]int{audit.IssueEmpty: 1},
	})

	// Verify it's valid JSON and contains expected fields
	if !strings.Contains(result, `"hasRisks":true`) {
		t.Error("expected hasRisks:true")
	}
	if !strings.Contains(result, `"type":"empty"`) {
		t.Error("expected type:empty")
	}
	if !strings.Contains(result, `"key":"DATABASE_URL"`) {
		t.Error("expected key:DATABASE_URL")
	}
	if !strings.Contains(result, `"message":"variable has empty value"`) {
		t.Error("expected message")
	}
}

// **Feature: env-audit-v2, Property 11: GitHub Actions format**
// **Validates: Requirements 9.1, 9.2**
// For any audit result, when --github flag is used, the output SHALL use
// ::error:: prefix for critical issues and ::warning:: prefix for non-critical issues.
func TestProperty_GitHubActionsFormat(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for issue type (0-5 covers all IssueType values)
	genIssueType := gen.IntRange(0, 5).Map(func(i int) audit.IssueType {
		return audit.IssueType(i)
	})

	// Generator for a single issue with non-empty alphanumeric key
	genIssue := gen.Struct(reflect.TypeOf(audit.Issue{}), map[string]gopter.Gen{
		"Type":    genIssueType,
		"Key":     gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		"Message": gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	})

	// Generator for slice of issues (at least 1)
	genIssues := gen.SliceOf(genIssue).SuchThat(func(issues []audit.Issue) bool {
		return len(issues) > 0
	})

	properties.Property("GitHub output uses correct prefixes for all issues", prop.ForAll(
		func(issues []audit.Issue) bool {
			result := &audit.Result{
				Issues:   issues,
				HasRisks: len(issues) > 0,
			}

			formatter := &GitHubFormatter{}
			output := formatter.Format(result)

			// Each line should start with ::error:: or ::warning::
			lines := strings.Split(output, "\n")
			for i, issue := range issues {
				if i >= len(lines) {
					return false
				}
				line := lines[i]

				// Critical issues (missing, leak, duplicate) use ::error::
				isCritical := issue.Type == audit.IssueMissing ||
					issue.Type == audit.IssueLeak ||
					issue.Type == audit.IssueDuplicate

				if isCritical {
					if !strings.HasPrefix(line, "::error::") {
						t.Logf("Expected ::error:: for type %d, got: %s", issue.Type, line)
						return false
					}
				} else {
					if !strings.HasPrefix(line, "::warning::") {
						t.Logf("Expected ::warning:: for type %d, got: %s", issue.Type, line)
						return false
					}
				}

				// Line should contain the key
				if !strings.Contains(line, issue.Key) {
					t.Logf("Line should contain key %s: %s", issue.Key, line)
					return false
				}
			}
			return true
		},
		genIssues,
	))

	properties.TestingRun(t)
}

func TestGitHubFormatter_NilResult(t *testing.T) {
	f := &GitHubFormatter{}
	result := f.Format(nil)
	if result != "" {
		t.Errorf("expected empty string for nil result, got %s", result)
	}
}

func TestGitHubFormatter_EmptyIssues(t *testing.T) {
	f := &GitHubFormatter{}
	result := f.Format(&audit.Result{
		Issues:   []audit.Issue{},
		HasRisks: false,
	})
	if result != "" {
		t.Errorf("expected empty string for empty issues, got %s", result)
	}
}

func TestGitHubFormatter_ErrorPrefix(t *testing.T) {
	f := &GitHubFormatter{}
	result := f.Format(&audit.Result{
		Issues: []audit.Issue{
			{Type: audit.IssueMissing, Key: "API_KEY", Message: "required variable is missing"},
			{Type: audit.IssueLeak, Key: "SECRET", Message: "potential leak detected"},
			{Type: audit.IssueDuplicate, Key: "DUPE", Message: "duplicate key"},
		},
		HasRisks: true,
	})

	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "::error::") {
			t.Errorf("expected ::error:: prefix for critical issue, got: %s", line)
		}
	}
}

func TestGitHubFormatter_WarningPrefix(t *testing.T) {
	f := &GitHubFormatter{}
	result := f.Format(&audit.Result{
		Issues: []audit.Issue{
			{Type: audit.IssueEmpty, Key: "EMPTY_VAR", Message: "variable has empty value"},
			{Type: audit.IssueSensitive, Key: "PASSWORD", Message: "sensitive key detected"},
			{Type: audit.IssueExtra, Key: "EXTRA", Message: "extra variable"},
		},
		HasRisks: true,
	})

	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "::warning::") {
			t.Errorf("expected ::warning:: prefix for non-critical issue, got: %s", line)
		}
	}
}

func TestGitHubFormatter_ContainsKeyAndMessage(t *testing.T) {
	f := &GitHubFormatter{}
	result := f.Format(&audit.Result{
		Issues: []audit.Issue{
			{Type: audit.IssueEmpty, Key: "MY_VAR", Message: "variable has empty value"},
		},
		HasRisks: true,
	})

	if !strings.Contains(result, "MY_VAR") {
		t.Error("output should contain the key")
	}
	if !strings.Contains(result, "variable has empty value") {
		t.Error("output should contain the message")
	}
}

// **Feature: env-audit-v2, Property 14: Color disabling**
// **Validates: Requirements 12.2, 12.3, 12.4**
// ShouldUseColor SHALL return false when --no-color flag is set,
// when NO_COLOR env var is set, or when stdout is not a TTY.
func TestProperty_ColorDisabling(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("--no-color flag disables color regardless of TTY", prop.ForAll(
		func(isTTY bool) bool {
			return !ShouldUseColor(true, isTTY)
		},
		gen.Bool(),
	))

	properties.Property("color disabled when not TTY", prop.ForAll(
		func(noColor bool) bool {
			if noColor {
				return true // skip, handled by other test
			}
			return !ShouldUseColor(false, false)
		},
		gen.Bool(),
	))

	properties.Property("color enabled only when TTY and no --no-color", prop.ForAll(
		func(_ bool) bool {
			return ShouldUseColor(false, true)
		},
		gen.Bool(),
	))

	properties.TestingRun(t)
}

func TestShouldUseColor_NoColorFlag(t *testing.T) {
	if ShouldUseColor(true, true) {
		t.Error("--no-color should disable color")
	}
	if ShouldUseColor(true, false) {
		t.Error("--no-color should disable color even if not TTY")
	}
}

func TestShouldUseColor_NotTTY(t *testing.T) {
	if ShouldUseColor(false, false) {
		t.Error("non-TTY should disable color")
	}
}

func TestShouldUseColor_TTY(t *testing.T) {
	if !ShouldUseColor(false, true) {
		t.Error("TTY without --no-color should enable color")
	}
}

func TestTextFormatter_NoIssues(t *testing.T) {
	f := &TextFormatter{UseColor: false}
	result := f.Format(&audit.Result{Issues: []audit.Issue{}})
	if !strings.Contains(result, "No issues found") {
		t.Error("expected 'No issues found'")
	}
}

func TestTextFormatter_NoIssuesWithColor(t *testing.T) {
	f := &TextFormatter{UseColor: true}
	result := f.Format(&audit.Result{Issues: []audit.Issue{}})
	if !strings.Contains(result, "\033[32m") {
		t.Error("expected green color code for success")
	}
	if !strings.Contains(result, "No issues found") {
		t.Error("expected 'No issues found'")
	}
}

func TestTextFormatter_WithIssues(t *testing.T) {
	f := &TextFormatter{UseColor: false}
	result := f.Format(&audit.Result{
		Issues: []audit.Issue{
			{Type: audit.IssueEmpty, Key: "EMPTY_VAR", Message: "variable has empty value"},
		},
		HasRisks: true,
	})
	if !strings.Contains(result, "EMPTY_VAR") {
		t.Error("expected EMPTY_VAR in output")
	}
	if !strings.Contains(result, "Empty Values") {
		t.Error("expected 'Empty Values' header")
	}
}

func TestTextFormatter_WithColor(t *testing.T) {
	f := &TextFormatter{UseColor: true}
	result := f.Format(&audit.Result{
		Issues: []audit.Issue{
			{Type: audit.IssueMissing, Key: "API_KEY", Message: "required variable is missing"},
		},
		HasRisks: true,
	})
	// Should contain red color for errors
	if !strings.Contains(result, "\033[31m") {
		t.Error("expected red color code for errors")
	}
	if !strings.Contains(result, "\033[0m") {
		t.Error("expected color reset code")
	}
}

func TestTextFormatter_WarningsYellow(t *testing.T) {
	f := &TextFormatter{UseColor: true}
	result := f.Format(&audit.Result{
		Issues: []audit.Issue{
			{Type: audit.IssueEmpty, Key: "EMPTY_VAR", Message: "variable has empty value"},
		},
		HasRisks: true,
	})
	// Should contain yellow color for warnings
	if !strings.Contains(result, "\033[33m") {
		t.Error("expected yellow color code for warnings")
	}
}