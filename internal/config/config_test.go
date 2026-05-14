package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withIsolatedEnv points HOME and XDG_CONFIG_HOME at temp dirs and clears
// TRY_PATH, so each test runs against a clean slate. The cleanup restores
// the previous environment.
func withIsolatedEnv(t *testing.T) (homeDir, xdgDir string) {
	t.Helper()
	homeDir = t.TempDir()
	xdgDir = t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", xdgDir)
	t.Setenv("TRY_PATH", "")
	return homeDir, xdgDir
}

func writeConfig(t *testing.T, xdgDir, body string) string {
	t.Helper()
	dir := filepath.Join(xdgDir, "try")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

func TestResolve_DefaultWhenNothingSet(t *testing.T) {
	homeDir, _ := withIsolatedEnv(t)

	r, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if r.Source != SourceDefault {
		t.Errorf("Source = %q, want %q", r.Source, SourceDefault)
	}
	want := filepath.Join(homeDir, "src", "tries")
	if r.Path != want {
		t.Errorf("Path = %q, want %q", r.Path, want)
	}
}

func TestResolve_ConfigFileBeatsDefault(t *testing.T) {
	homeDir, xdgDir := withIsolatedEnv(t)
	writeConfig(t, xdgDir, `tries_path = "~/custom-tries"`+"\n")

	r, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if r.Source != SourceConfigFile {
		t.Errorf("Source = %q, want %q", r.Source, SourceConfigFile)
	}
	want := filepath.Join(homeDir, "custom-tries")
	if r.Path != want {
		t.Errorf("Path = %q, want %q", r.Path, want)
	}
}

func TestResolve_EnvBeatsConfigFile(t *testing.T) {
	homeDir, xdgDir := withIsolatedEnv(t)
	writeConfig(t, xdgDir, `tries_path = "~/from-file"`+"\n")
	t.Setenv("TRY_PATH", filepath.Join(homeDir, "from-env"))

	r, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if r.Source != SourceEnv {
		t.Errorf("Source = %q, want %q", r.Source, SourceEnv)
	}
	want := filepath.Join(homeDir, "from-env")
	if r.Path != want {
		t.Errorf("Path = %q, want %q", r.Path, want)
	}
}

func TestResolve_BlankConfigValueFallsThroughToDefault(t *testing.T) {
	_, xdgDir := withIsolatedEnv(t)
	// Empty / whitespace tries_path should not win over default.
	writeConfig(t, xdgDir, `tries_path = "   "`+"\n")

	r, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if r.Source != SourceDefault {
		t.Errorf("Source = %q, want %q", r.Source, SourceDefault)
	}
}

func TestResolve_MalformedConfigSurfacesError(t *testing.T) {
	_, xdgDir := withIsolatedEnv(t)
	writeConfig(t, xdgDir, "tries_path = [not, a, string]\n")

	_, err := Resolve()
	if err == nil {
		t.Fatal("expected error for malformed config, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("error %q does not mention parse failure", err)
	}
}

// TestResolve_MalformedConfigShortCircuitsBeforeEnv pins current behavior:
// if the config file is unparseable, we surface the error rather than
// silently falling through to $TRY_PATH. The reasoning is that a broken
// config is almost always a bug the user wants to know about, and silently
// ignoring it would mask the problem.
func TestResolve_MalformedConfigShortCircuitsBeforeEnv(t *testing.T) {
	homeDir, xdgDir := withIsolatedEnv(t)
	writeConfig(t, xdgDir, "this is not valid toml = = =\n")
	t.Setenv("TRY_PATH", filepath.Join(homeDir, "from-env"))

	_, err := Resolve()
	if err == nil {
		t.Fatal("expected error from malformed config even when env is set, got nil")
	}
}

// TestResolve_RejectsShellMetacharsInConfig and ...InEnv pin the security fix:
// values containing NUL or newline are rejected at config-load time, even
// before they would be shell-quoted at init-script generation.
func TestResolve_RejectsShellMetacharsInConfig(t *testing.T) {
	_, xdgDir := withIsolatedEnv(t)
	writeConfig(t, xdgDir, "tries_path = \"/tmp/x\\nrm -rf /\"\n")

	_, err := Resolve()
	if err == nil {
		t.Fatal("expected error for newline in config tries_path, got nil")
	}
	if !strings.Contains(err.Error(), "newline") {
		t.Errorf("error %q should mention newline/NUL", err)
	}
}

func TestResolve_RejectsShellMetacharsInEnv(t *testing.T) {
	withIsolatedEnv(t)
	t.Setenv("TRY_PATH", "/tmp/x\nrm -rf /")

	_, err := Resolve()
	if err == nil {
		t.Fatal("expected error for newline in $TRY_PATH, got nil")
	}
}

func TestGetBasePath_CreatesDir(t *testing.T) {
	homeDir, _ := withIsolatedEnv(t)
	t.Setenv("TRY_PATH", filepath.Join(homeDir, "made-on-demand"))

	got, err := GetBasePath()
	if err != nil {
		t.Fatalf("GetBasePath: %v", err)
	}
	want := filepath.Join(homeDir, "made-on-demand")
	if got != want {
		t.Errorf("path = %q, want %q", got, want)
	}
	if info, err := os.Stat(got); err != nil || !info.IsDir() {
		t.Errorf("directory was not created at %s (err=%v)", got, err)
	}
}

func TestConfigFilePath_HonorsXDG(t *testing.T) {
	_, xdgDir := withIsolatedEnv(t)
	got, err := ConfigFilePath()
	if err != nil {
		t.Fatalf("ConfigFilePath: %v", err)
	}
	want := filepath.Join(xdgDir, "try", "config.toml")
	if got != want {
		t.Errorf("ConfigFilePath = %q, want %q", got, want)
	}
}

func TestConfigFilePath_FallsBackToHomeDotConfig(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", "")

	got, err := ConfigFilePath()
	if err != nil {
		t.Fatalf("ConfigFilePath: %v", err)
	}
	want := filepath.Join(homeDir, ".config", "try", "config.toml")
	if got != want {
		t.Errorf("ConfigFilePath = %q, want %q", got, want)
	}
}
