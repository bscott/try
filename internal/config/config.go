package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	defaultTriesDir = "~/src/tries"
	envVarName      = "TRY_PATH"
	configFileName  = "config.toml"
	configSubDir    = "try"

	// maxConfigSize caps how much of the config file we will parse. The file
	// is fully user-controlled, but a runaway value (e.g. accidentally
	// concatenated logs) shouldn't OOM the CLI.
	maxConfigSize = 64 * 1024
)

// shellMetaChars rejects values that would be unsafe to splice into a shell
// init script even after quoting (NUL truncates C strings; CR/LF break the
// generated script's line structure). Single quotes are *not* rejected — the
// shell.shellQuote helper escapes them correctly.
const shellMetaChars = "\x00\r\n"

// Source identifies where the effective tries path came from.
type Source string

const (
	SourceDefault    Source = "default"
	SourceConfigFile Source = "config"
	SourceEnv        Source = "env"
)

// Resolved is the result of resolving the tries path, including which source won.
type Resolved struct {
	Path       string // Absolute, ~-expanded, cleaned
	Source     Source
	ConfigFile string // Path the loader would consult, regardless of whether the file exists
	EnvVar     string // Name of the env var consulted (always "TRY_PATH" today)
	Promote    Promote
}

// Promote holds resolved settings for the `try promote` subcommand.
// All fields are populated with defaults when the user hasn't set them.
type Promote struct {
	Root   string // Absolute, ~-expanded, cleaned. Defaults to parent of tries Path.
	Depth  int    // >=1. Defaults to DefaultPromoteDepth.
	Picker string // Defaults to DefaultPromotePicker.
}

// fileConfig is the on-disk schema for ~/.config/try/config.toml. Kept
// unexported — callers should not write config programmatically; they
// should hand-edit the file or use Resolve to read it back.
type fileConfig struct {
	TriesPath string         `toml:"tries_path"`
	Promote   promoteSection `toml:"promote"`
}

// promoteSection is the [promote] table — settings for the `try promote`
// subcommand. All fields are optional; empty values fall back to defaults
// computed in Resolve.
type promoteSection struct {
	Root   string `toml:"root"`   // Root dir to fuzzy-pick destinations from. Default: parent of tries_path.
	Depth  int    `toml:"depth"`  // Directory walk depth under Root for candidate destinations. Default: 1.
	Picker string `toml:"picker"` // Picker tool. Default: "fzf". Today only "fzf" is supported.
}

// Default values for the [promote] section. Exposed so callers and tests
// can compare against them.
const (
	DefaultPromoteDepth  = 1
	DefaultPromotePicker = "fzf"
)

// ConfigFilePath returns the path the config loader will read from.
// It honors $XDG_CONFIG_HOME, falling back to ~/.config.
func ConfigFilePath() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, configSubDir, configFileName), nil
}

// loadFileConfig reads the config file if it exists. A missing file is not
// an error. The file is read with a size cap (maxConfigSize) so a runaway
// file can't OOM the CLI.
func loadFileConfig() (fileConfig, string, error) {
	path, err := ConfigFilePath()
	if err != nil {
		return fileConfig{}, "", err
	}

	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fileConfig{}, path, nil
		}
		return fileConfig{}, path, fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer f.Close()

	var fc fileConfig
	if _, err := toml.NewDecoder(io.LimitReader(f, maxConfigSize)).Decode(&fc); err != nil {
		return fileConfig{}, path, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return fc, path, nil
}

// Resolve computes the effective tries path with full precedence:
//
//	env (TRY_PATH) > config file (tries_path) > default (~/src/tries)
//
// The returned Path is absolute, ~-expanded, and cleaned. The directory is NOT
// created here — call EnsureDir on Resolved.Path if you need the dir to exist.
func Resolve() (Resolved, error) {
	r := Resolved{
		Source: SourceDefault,
		EnvVar: envVarName,
		Path:   defaultTriesDir,
	}

	fc, configPath, err := loadFileConfig()
	r.ConfigFile = configPath
	if err != nil {
		return r, err
	}
	if v := strings.TrimSpace(fc.TriesPath); v != "" {
		if err := validateRawPath(v, "config file "+configPath); err != nil {
			return r, err
		}
		r.Path = v
		r.Source = SourceConfigFile
	}

	if env := os.Getenv(envVarName); env != "" {
		if err := validateRawPath(env, "$"+envVarName); err != nil {
			return r, err
		}
		r.Path = env
		r.Source = SourceEnv
	}

	expanded, err := expandPath(r.Path)
	if err != nil {
		return r, err
	}
	r.Path = expanded

	promote, err := resolvePromote(fc.Promote, r.Path)
	if err != nil {
		return r, err
	}
	r.Promote = promote
	return r, nil
}

// resolvePromote fills in defaults for the [promote] section. `triesPath` is
// the already-resolved (absolute, expanded) tries directory — used to derive
// the default Root (its parent).
func resolvePromote(fc promoteSection, triesPath string) (Promote, error) {
	p := Promote{
		Root:   strings.TrimSpace(fc.Root),
		Depth:  fc.Depth,
		Picker: strings.TrimSpace(fc.Picker),
	}
	if p.Root == "" {
		p.Root = filepath.Dir(triesPath)
	} else {
		if err := validateRawPath(p.Root, "config file [promote].root"); err != nil {
			return p, err
		}
		expanded, err := expandPath(p.Root)
		if err != nil {
			return p, err
		}
		p.Root = expanded
	}
	if p.Depth <= 0 {
		p.Depth = DefaultPromoteDepth
	}
	if p.Picker == "" {
		p.Picker = DefaultPromotePicker
	}
	return p, nil
}

// validateRawPath rejects values that we know are unsafe to splice into the
// generated shell init script, even after quoting. We intentionally do NOT
// reject ".." or absolute paths: the user owns their own config and may
// legitimately point the tries dir anywhere they can write.
func validateRawPath(p, origin string) error {
	if strings.ContainsAny(p, shellMetaChars) {
		return fmt.Errorf("tries path from %s contains a NUL or newline character, refusing to use it", origin)
	}
	return nil
}

// GetBasePath preserves the original API: returns the effective tries dir as an
// absolute path, creating it if it doesn't exist. Callers that want the source
// info should use Resolve() instead.
func GetBasePath() (string, error) {
	r, err := Resolve()
	if err != nil {
		return "", err
	}
	if err := EnsureDir(r.Path); err != nil {
		return "", err
	}
	return r.Path, nil
}

// EnsureDir creates the directory if it doesn't already exist.
//
// We deliberately use os.MkdirAll here (and not os.Mkdir) because this is
// the *base* tries directory, which may be nested under paths the user
// hasn't created yet (e.g. ~/code/tries before ~/code exists). The CLAUDE.md
// "use os.Mkdir" guidance applies to per-entry directories (each new try),
// where TOCTOU between exists-check and create matters.
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("failed to create tries directory: %w", err)
	}
	return nil
}

// expandPath turns "~/foo" into "/home/user/foo" and cleans the result.
func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		if path == "~" {
			path = home
		} else {
			path = filepath.Join(home, path[2:])
		}
	}
	return filepath.Clean(path), nil
}
