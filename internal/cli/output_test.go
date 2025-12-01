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
