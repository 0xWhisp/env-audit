package cli

import "testing"

// Unit tests for CLI argument parsing
// _Requirements: 4.4, 5.3_

func TestParseArgs_ValidArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Config
	}{
		{
			name:     "help flag long",
			args:     []string{"--help"},
			expected: Config{Help: true},
		},
		{
			name:     "help flag short",
			args:     []string{"-h"},
			expected: Config{Help: true},
		},
		{
			name:     "dump flag long",
			args:     []string{"--dump"},
			expected: Config{DumpMode: true},
		},
		{
			name:     "dump flag short",
			args:     []string{"-d"},
			expected: Config{DumpMode: true},
		},
		{
			name:     "json flag",
			args:     []string{"--json"},
			expected: Config{JSONOutput: true},
		},
		{
			name:     "quiet flag long",
			args:     []string{"--quiet"},
			expected: Config{Quiet: true},
		},
		{
			name:     "quiet flag short",
			args:     []string{"-q"},
			expected: Config{Quiet: true},
		},
		{
			name:     "strict flag",
			args:     []string{"--strict"},
			expected: Config{Strict: true},
		},
		{
			name:     "check-leaks flag",
			args:     []string{"--check-leaks"},
			expected: Config{CheckLeaks: true},
		},
		{
			name:     "init flag",
			args:     []string{"--init"},
			expected: Config{Init: true},
		},
		{
			name:     "force flag",
			args:     []string{"--force"},
			expected: Config{Force: true},
		},
		{
			name:     "init with force",
			args:     []string{"--init", "--force"},
			expected: Config{Init: true, Force: true},
		},
		{
			name:     "file flag long",
			args:     []string{"--file", ".env"},
			expected: Config{FilePath: ".env"},
		},
		{
			name:     "file flag short",
			args:     []string{"-f", "config.env"},
			expected: Config{FilePath: "config.env"},
		},
		{
			name:     "required flag long",
			args:     []string{"--required", "VAR1,VAR2"},
			expected: Config{Required: []string{"VAR1", "VAR2"}},
		},
		{
			name:     "required flag short",
			args:     []string{"-r", "API_KEY"},
			expected: Config{Required: []string{"API_KEY"}},
		},
		{
			name:     "multiple flags combined",
			args:     []string{"-f", "test.env", "-r", "A,B", "-d"},
			expected: Config{FilePath: "test.env", Required: []string{"A", "B"}, DumpMode: true},
		},
		{
			name:     "no args",
			args:     []string{},
			expected: Config{},
		},
		{
			name:     "diff flag",
			args:     []string{"--diff", "other.env"},
			expected: Config{DiffFile: "other.env"},
		},
		{
			name:     "diff with file",
			args:     []string{"-f", ".env", "--diff", "prod.env"},
			expected: Config{FilePath: ".env", DiffFile: "prod.env"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Help != tt.expected.Help {
				t.Errorf("Help: got %v, want %v", cfg.Help, tt.expected.Help)
			}
			if cfg.DumpMode != tt.expected.DumpMode {
				t.Errorf("DumpMode: got %v, want %v", cfg.DumpMode, tt.expected.DumpMode)
			}
			if cfg.JSONOutput != tt.expected.JSONOutput {
				t.Errorf("JSONOutput: got %v, want %v", cfg.JSONOutput, tt.expected.JSONOutput)
			}
			if cfg.Quiet != tt.expected.Quiet {
				t.Errorf("Quiet: got %v, want %v", cfg.Quiet, tt.expected.Quiet)
			}
			if cfg.Strict != tt.expected.Strict {
				t.Errorf("Strict: got %v, want %v", cfg.Strict, tt.expected.Strict)
			}
			if cfg.CheckLeaks != tt.expected.CheckLeaks {
				t.Errorf("CheckLeaks: got %v, want %v", cfg.CheckLeaks, tt.expected.CheckLeaks)
			}
			if cfg.Init != tt.expected.Init {
				t.Errorf("Init: got %v, want %v", cfg.Init, tt.expected.Init)
			}
			if cfg.Force != tt.expected.Force {
				t.Errorf("Force: got %v, want %v", cfg.Force, tt.expected.Force)
			}
			if cfg.FilePath != tt.expected.FilePath {
				t.Errorf("FilePath: got %v, want %v", cfg.FilePath, tt.expected.FilePath)
			}
			if cfg.DiffFile != tt.expected.DiffFile {
				t.Errorf("DiffFile: got %v, want %v", cfg.DiffFile, tt.expected.DiffFile)
			}
			if len(cfg.Required) != len(tt.expected.Required) {
				t.Errorf("Required length: got %v, want %v", len(cfg.Required), len(tt.expected.Required))
			}
			for i := range cfg.Required {
				if cfg.Required[i] != tt.expected.Required[i] {
					t.Errorf("Required[%d]: got %v, want %v", i, cfg.Required[i], tt.expected.Required[i])
				}
			}
		})
	}
}


func TestParseArgs_InvalidArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "unknown flag", args: []string{"--unknown"}},
		{name: "missing file value", args: []string{"--file"}},
		{name: "missing file value short", args: []string{"-f"}},
		{name: "missing required value", args: []string{"--required"}},
		{name: "missing required value short", args: []string{"-r"}},
		{name: "missing diff value", args: []string{"--diff"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseArgs(tt.args)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestParseCommaSeparated_Empty(t *testing.T) {
	result := parseCommaSeparated("")
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestTrimSpace_AllSpaces(t *testing.T) {
	result := trimSpace("   ")
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestTrimSpace_Tabs(t *testing.T) {
	result := trimSpace("\t\tvalue\t\t")
	if result != "value" {
		t.Errorf("expected 'value', got %q", result)
	}
}
