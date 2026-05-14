package config

import (
	"errors"
	"fmt"
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
)

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
	ConfigFile string // Path to the config file consulted (empty if none / not loaded)
	EnvVar     string // Name of the env var consulted ("" if none)
}

// FileConfig is the on-disk schema for ~/.config/try/config.toml.
type FileConfig struct {
	TriesPath string `toml:"tries_path"`
}

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

// loadFileConfig reads the config file if it exists. A missing file is not an error.
func loadFileConfig() (FileConfig, string, error) {
	path, err := ConfigFilePath()
	if err != nil {
		return FileConfig{}, "", err
	}
	var fc FileConfig
	_, err = toml.DecodeFile(path, &fc)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return FileConfig{}, path, nil
		}
		return FileConfig{}, path, fmt.Errorf("failed to parse %s: %w", path, err)
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
	if strings.TrimSpace(fc.TriesPath) != "" {
		r.Path = fc.TriesPath
		r.Source = SourceConfigFile
	}

	if env := os.Getenv(envVarName); env != "" {
		r.Path = env
		r.Source = SourceEnv
	}

	expanded, err := expandPath(r.Path)
	if err != nil {
		return r, err
	}
	r.Path = expanded
	return r, nil
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
