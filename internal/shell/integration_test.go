package shell

import (
	"strings"
	"testing"
)

func TestDetectShell(t *testing.T) {
	tests := []struct {
		name     string
		shellEnv string
		want     ShellType
	}{
		{
			name:     "bash full path",
			shellEnv: "/bin/bash",
			want:     ShellBash,
		},
		{
			name:     "bash usr bin",
			shellEnv: "/usr/bin/bash",
			want:     ShellBash,
		},
		{
			name:     "zsh full path",
			shellEnv: "/bin/zsh",
			want:     ShellZsh,
		},
		{
			name:     "zsh usr local",
			shellEnv: "/usr/local/bin/zsh",
			want:     ShellZsh,
		},
		{
			name:     "fish full path",
			shellEnv: "/usr/bin/fish",
			want:     ShellFish,
		},
		{
			name:     "unknown shell",
			shellEnv: "/bin/sh",
			want:     ShellUnknown,
		},
		{
			name:     "empty string",
			shellEnv: "",
			want:     ShellUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectShell(tt.shellEnv)
			if got != tt.want {
				t.Errorf("DetectShell(%q) = %v, want %v", tt.shellEnv, got, tt.want)
			}
		})
	}
}

func TestGenerateBashInit(t *testing.T) {
	basePath := "/home/user/tries"
	result := GenerateBashInit(basePath)

	// Check key components
	if !strings.Contains(result, "try()") {
		t.Error("Expected function definition 'try()'")
	}
	if !strings.Contains(result, "command try exec") {
		t.Error("Expected 'command try exec' call")
	}
	if !strings.Contains(result, "2>/dev/tty") {
		t.Error("Expected stderr redirection to /dev/tty")
	}
	if !strings.Contains(result, "eval \"$result\"") {
		t.Error("Expected eval of result")
	}
	if !strings.Contains(result, basePath) {
		t.Errorf("Expected base path %s in output", basePath)
	}
	if !strings.Contains(result, "export TRY_PATH=") {
		t.Error("Expected TRY_PATH export")
	}
}

func TestGenerateZshInit(t *testing.T) {
	basePath := "/home/user/tries"
	result := GenerateZshInit(basePath)

	// Check key components (same as bash)
	if !strings.Contains(result, "try()") {
		t.Error("Expected function definition 'try()'")
	}
	if !strings.Contains(result, "command try exec") {
		t.Error("Expected 'command try exec' call")
	}
	if !strings.Contains(result, "2>/dev/tty") {
		t.Error("Expected stderr redirection to /dev/tty")
	}
	if !strings.Contains(result, "eval \"$result\"") {
		t.Error("Expected eval of result")
	}
	if !strings.Contains(result, basePath) {
		t.Errorf("Expected base path %s in output", basePath)
	}
	if !strings.Contains(result, "export TRY_PATH=") {
		t.Error("Expected TRY_PATH export")
	}
}

func TestGenerateFishInit(t *testing.T) {
	basePath := "/home/user/tries"
	result := GenerateFishInit(basePath)

	// Check key components (fish syntax)
	if !strings.Contains(result, "function try") {
		t.Error("Expected function definition 'function try'")
	}
	if !strings.Contains(result, "command try exec") {
		t.Error("Expected 'command try exec' call")
	}
	if !strings.Contains(result, "2>/dev/tty") {
		t.Error("Expected stderr redirection to /dev/tty")
	}
	if !strings.Contains(result, "eval $result") {
		t.Error("Expected eval of result")
	}
	if !strings.Contains(result, "test $status -eq 0") {
		t.Error("Expected status check")
	}
	if !strings.Contains(result, basePath) {
		t.Errorf("Expected base path %s in output", basePath)
	}
	if !strings.Contains(result, "set -gx TRY_PATH") {
		t.Error("Expected TRY_PATH export")
	}
}

// TestGenerateInit_QuotesMaliciousBasePath pins the security fix that a
// hostile basePath (from $TRY_PATH or a config file) cannot inject shell
// commands into the generated init script.
func TestGenerateInit_QuotesMaliciousBasePath(t *testing.T) {
	// A value with `"; ...; #` would, without quoting, escape the
	// surrounding double-quoted assignment and inject a command.
	injected := `/tmp/x"; touch /tmp/PWNED; #`

	for _, shell := range []ShellType{ShellBash, ShellZsh, ShellFish} {
		got, err := GenerateInit(shell, injected)
		if err != nil {
			t.Fatalf("GenerateInit(%v): %v", shell, err)
		}
		// The unquoted suffix must not appear in the output; the whole
		// value should be inside a single-quoted string.
		if strings.Contains(got, `touch /tmp/PWNED`) && !strings.Contains(got, `'/tmp/x"; touch /tmp/PWNED; #'`) {
			t.Errorf("shell %v: malicious value not quoted, output:\n%s", shell, got)
		}
		// A bare unquoted assignment is the smoking gun.
		if strings.Contains(got, `TRY_PATH="`+injected) {
			t.Errorf("shell %v: basePath appears inside double quotes (unquoted form), output:\n%s", shell, got)
		}
	}
}

// TestGenerateInit_QuotesSingleQuoteInBasePath covers paths with a literal
// single quote, which is the awkward case for the standard '\'' escape.
func TestGenerateInit_QuotesSingleQuoteInBasePath(t *testing.T) {
	val := `/tmp/o'reilly`
	got := GenerateBashInit(val)
	want := `'/tmp/o'\''reilly'`
	if !strings.Contains(got, want) {
		t.Errorf("expected escaped form %s in output, got:\n%s", want, got)
	}
}

func TestGenerateInit(t *testing.T) {
	basePath := "/home/user/tries"

	tests := []struct {
		name      string
		shellType ShellType
		wantErr   bool
		contains  string
	}{
		{
			name:      "bash",
			shellType: ShellBash,
			wantErr:   false,
			contains:  "try()",
		},
		{
			name:      "zsh",
			shellType: ShellZsh,
			wantErr:   false,
			contains:  "try()",
		},
		{
			name:      "fish",
			shellType: ShellFish,
			wantErr:   false,
			contains:  "function try",
		},
		{
			name:      "unknown",
			shellType: ShellUnknown,
			wantErr:   true,
			contains:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateInit(tt.shellType, basePath)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !strings.Contains(got, tt.contains) {
				t.Errorf("Expected output to contain %q", tt.contains)
			}
		})
	}
}
