package audit

import "testing"

func TestScan_NoIssues(t *testing.T) {
	env := map[string]string{"APP_NAME": "test"}
	result := Scan(env, nil)

	if result.HasRisks {
		t.Error("expected no risks")
	}
	if len(result.Issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(result.Issues))
	}
}

func TestScan_EmptyValues(t *testing.T) {
	env := map[string]string{"DB_URL": ""}
	result := Scan(env, nil)

	// Empty values are warnings, not risks (unless strict mode)
	if result.HasRisks {
		t.Error("expected no risks for warnings without strict mode")
	}
	if len(result.Issues) != 1 || result.Issues[0].Type != IssueEmpty {
		t.Errorf("expected 1 empty issue, got %v", result.Issues)
	}
}

func TestScan_EmptyValues_Strict(t *testing.T) {
	env := map[string]string{"DB_URL": ""}
	result := Scan(env, &ScanOptions{Strict: true})

	// In strict mode, warnings become risks
	if !result.HasRisks {
		t.Error("expected risks in strict mode")
	}
	if len(result.Issues) != 1 || result.Issues[0].Type != IssueEmpty {
		t.Errorf("expected 1 empty issue, got %v", result.Issues)
	}
}

func TestScan_MissingRequired(t *testing.T) {
	env := map[string]string{"FOO": "bar"}
	result := Scan(env, &ScanOptions{Required: []string{"MISSING"}})

	if !result.HasRisks {
		t.Error("expected risks")
	}
	if len(result.Issues) != 1 || result.Issues[0].Type != IssueMissing {
		t.Errorf("expected 1 missing issue, got %v", result.Issues)
	}
}

func TestScan_SensitiveKeys(t *testing.T) {
	env := map[string]string{"API_SECRET": "hidden"}
	result := Scan(env, nil)

	// Sensitive keys are info-level, never risks
	if result.HasRisks {
		t.Error("expected no risks for info-level issues")
	}
	if len(result.Issues) != 1 || result.Issues[0].Type != IssueSensitive {
		t.Errorf("expected 1 sensitive issue, got %v", result.Issues)
	}
}

func TestScan_Duplicates(t *testing.T) {
	env := map[string]string{"FOO": "bar"}
	result := Scan(env, &ScanOptions{Duplicates: []string{"FOO"}})

	// Duplicates are warnings, not risks (unless strict mode)
	if result.HasRisks {
		t.Error("expected no risks for warnings without strict mode")
	}
	if len(result.Issues) != 1 || result.Issues[0].Type != IssueDuplicate {
		t.Errorf("expected 1 duplicate issue, got %v", result.Issues)
	}
}

func TestScan_Duplicates_Strict(t *testing.T) {
	env := map[string]string{"FOO": "bar"}
	result := Scan(env, &ScanOptions{Duplicates: []string{"FOO"}, Strict: true})

	// In strict mode, warnings become risks
	if !result.HasRisks {
		t.Error("expected risks in strict mode")
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
	result := Scan(env, &ScanOptions{
		Required:   []string{"MISSING"},
		Duplicates: []string{"DUP"},
	})

	if !result.HasRisks {
		t.Error("expected risks")
	}
	if len(result.Issues) != 4 {
		t.Errorf("expected 4 issues, got %d", len(result.Issues))
	}
}

func TestScan_WithIgnore(t *testing.T) {
	env := map[string]string{
		"EMPTY_VAR":  "",
		"API_SECRET": "val",
	}
	result := Scan(env, &ScanOptions{
		Ignore: []string{"EMPTY_VAR", "API_SECRET"},
	})

	if result.HasRisks {
		t.Error("expected no risks when all keys ignored")
	}
	if len(result.Issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(result.Issues))
	}
}

func TestScan_Summary(t *testing.T) {
	env := map[string]string{
		"EMPTY1": "",
		"EMPTY2": "",
		"SECRET": "val",
	}
	result := Scan(env, &ScanOptions{Required: []string{"MISSING"}})

	if result.Summary[IssueEmpty] != 2 {
		t.Errorf("expected 2 empty in summary, got %d", result.Summary[IssueEmpty])
	}
	if result.Summary[IssueMissing] != 1 {
		t.Errorf("expected 1 missing in summary, got %d", result.Summary[IssueMissing])
	}
	if result.Summary[IssueSensitive] != 1 {
		t.Errorf("expected 1 sensitive in summary, got %d", result.Summary[IssueSensitive])
	}
}
