package cli

import (
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
