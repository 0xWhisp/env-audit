package parser

import (
	"os"
	"testing"
)

func TestReadOSEnv(t *testing.T) {
	// Set a test env var
	os.Setenv("TEST_ENV_AUDIT_VAR", "testvalue")
	defer os.Unsetenv("TEST_ENV_AUDIT_VAR")

	env := ReadOSEnv()

	if env["TEST_ENV_AUDIT_VAR"] != "testvalue" {
		t.Errorf("expected testvalue, got %s", env["TEST_ENV_AUDIT_VAR"])
	}
}

func TestReadOSEnv_EmptyValue(t *testing.T) {
	os.Setenv("TEST_ENV_AUDIT_EMPTY", "")
	defer os.Unsetenv("TEST_ENV_AUDIT_EMPTY")

	env := ReadOSEnv()

	if val, exists := env["TEST_ENV_AUDIT_EMPTY"]; !exists || val != "" {
		t.Errorf("expected empty string, got %q, exists=%v", val, exists)
	}
}
