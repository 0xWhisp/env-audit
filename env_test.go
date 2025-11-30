package main

import (
	"os"
	"testing"
)

func TestReadEnv(t *testing.T) {
	// Set a test env var
	os.Setenv("TEST_ENV_AUDIT_VAR", "testvalue")
	defer os.Unsetenv("TEST_ENV_AUDIT_VAR")

	env := ReadEnv()

	if env["TEST_ENV_AUDIT_VAR"] != "testvalue" {
		t.Errorf("expected testvalue, got %s", env["TEST_ENV_AUDIT_VAR"])
	}
}

func TestReadEnv_EmptyValue(t *testing.T) {
	os.Setenv("TEST_ENV_AUDIT_EMPTY", "")
	defer os.Unsetenv("TEST_ENV_AUDIT_EMPTY")

	env := ReadEnv()

	if val, exists := env["TEST_ENV_AUDIT_EMPTY"]; !exists || val != "" {
		t.Errorf("expected empty string, got %q, exists=%v", val, exists)
	}
}
