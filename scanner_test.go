package main

import "testing"

func TestScan_NoIssues(t *testing.T) {
	env := map[string]string{"APP_NAME": "test"}
	result := Scan(env, nil, nil)

	if result.HasRisks {
		t.Error("expected no risks")
	}
	if len(result.Issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(result.Issues))
	}
}

func TestScan_EmptyValues(t *testing.T) {
	env := map[string]string{"DB_URL": ""}
	result := Scan(env, nil, nil)

	if !result.HasRisks {
		t.Error("expected risks")
	}
	if len(result.Issues) != 1 || result.Issues[0].Type != IssueEmpty {
		t.Errorf("expected 1 empty issue, got %v", result.Issues)
	}
}

func TestScan_MissingRequired(t *testing.T) {
	env := map[string]string{"FOO": "bar"}
	result := Scan(env, []string{"MISSING"}, nil)

	if !result.HasRisks {
		t.Error("expected risks")
	}
	if len(result.Issues) != 1 || result.Issues[0].Type != IssueMissing {
		t.Errorf("expected 1 missing issue, got %v", result.Issues)
	}
}

func TestScan_SensitiveKeys(t *testing.T) {
	env := map[string]string{"API_SECRET": "hidden"}
	result := Scan(env, nil, nil)

	if !result.HasRisks {
		t.Error("expected risks")
	}
	if len(result.Issues) != 1 || result.Issues[0].Type != IssueSensitive {
		t.Errorf("expected 1 sensitive issue, got %v", result.Issues)
	}
}

func TestScan_Duplicates(t *testing.T) {
	env := map[string]string{"FOO": "bar"}
	result := Scan(env, nil, []string{"FOO"})

	if !result.HasRisks {
		t.Error("expected risks")
	}
	if len(result.Issues) != 1 || result.Issues[0].Type != IssueDuplicate {
		t.Errorf("expected 1 duplicate issue, got %v", result.Issues)
	}
}

func TestScan_AllIssueTypes(t *testing.T) {
	env := map[string]string{
		"EMPTY_VAR":  "",
		"API_SECRET": "val",
	}
	result := Scan(env, []string{"MISSING"}, []string{"DUP"})

	if !result.HasRisks {
		t.Error("expected risks")
	}
	if len(result.Issues) != 4 {
		t.Errorf("expected 4 issues, got %d", len(result.Issues))
	}
}
