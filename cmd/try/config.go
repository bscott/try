package main

import (
	"fmt"

	"github.com/bscott/try/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View try configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the effective tries path and where it came from",
	Long: `Show the effective tries path resolved from (highest precedence first):

  1. $TRY_PATH environment variable
  2. tries_path in ~/.config/try/config.toml (or $XDG_CONFIG_HOME/try/config.toml)
  3. The compiled-in default (~/src/tries)
`,
	RunE: runConfigShow,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	r, err := config.Resolve()
	if err != nil {
		return err
	}
	fmt.Printf("tries_path  : %s\n", r.Path)
	fmt.Printf("source      : %s\n", r.Source)
	fmt.Printf("config file : %s\n", r.ConfigFile)
	fmt.Printf("env var     : %s\n", r.EnvVar)
	return nil
}
