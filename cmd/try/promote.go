package main

import (
	"fmt"
	"os"

	"github.com/bscott/try/internal/config"
	"github.com/bscott/try/internal/promote"
	"github.com/spf13/cobra"
)

var promoteCmd = &cobra.Command{
	Use:   "promote [try-name]",
	Short: "Move a try directory into ~/code (or wherever promote.root points)",
	Long: `Graduate a directory out of the tries dir into a real project location.

The promote workflow:
  1. Pick which try to graduate (interactively if [try-name] is omitted).
  2. Pick the destination parent directory under promote.root (interactively,
     unless --to is given).
  3. Move the dir, stripping the YYYY-MM-DD- date prefix (unless --keep-date).

Interactive selection uses fzf when it's installed. If fzf isn't on PATH,
promote falls back to a built-in fuzzy picker automatically, so it works
out-of-the-box. Set picker = "builtin" to always use the built-in picker.

The new path is also emitted to stdout as a "cd <path>" line, so if your
shell wrapper evals try's stdout (see ` + "`try init`" + `), promoting auto-cd's
you into the new location.

Configure defaults in ~/.config/try/config.toml:

  [promote]
  root   = "~/code"   # parent of tries_path by default
  depth  = 1          # directory walk depth for destination candidates
  picker = "fzf"      # "fzf" (falls back to built-in if fzf is absent) or "builtin"

Examples:
  try promote                                  # fully interactive
  try promote 2026-02-11-subtrackr-website     # named source, pick dest
  try promote subtrackr-website --to personal  # named source, named dest
  try promote --dry-run                        # show planned move only
`,
	RunE: runPromote,
}

func init() {
	rootCmd.AddCommand(promoteCmd)

	promoteCmd.Flags().StringP("to", "t", "", "Destination dir under promote.root (skip dest picker). Relative or absolute.")
	promoteCmd.Flags().String("root", "", "Override [promote].root for this invocation.")
	promoteCmd.Flags().Int("depth", 0, "Override [promote].depth for this invocation.")
	promoteCmd.Flags().Bool("keep-date", false, "Preserve the YYYY-MM-DD- date prefix in the new dir name.")
	promoteCmd.Flags().Bool("dry-run", false, "Describe the planned move without performing it.")
}

func runPromote(cmd *cobra.Command, args []string) error {
	cfg, err := config.Resolve()
	if err != nil {
		return fmt.Errorf("resolve config: %w", err)
	}
	if err := config.EnsureDir(cfg.Path); err != nil {
		return fmt.Errorf("ensure tries dir: %w", err)
	}

	opts := promote.Options{}
	if len(args) > 0 {
		opts.SourceName = args[0]
	}
	opts.DestDir, _ = cmd.Flags().GetString("to")
	opts.DestRoot, _ = cmd.Flags().GetString("root")
	opts.MaxDepth, _ = cmd.Flags().GetInt("depth")
	opts.KeepDate, _ = cmd.Flags().GetBool("keep-date")
	opts.DryRun, _ = cmd.Flags().GetBool("dry-run")

	res, err := promote.Run(cfg, opts)
	if err != nil {
		return err
	}

	// Status line goes to stderr so it doesn't pollute the cd command on
	// stdout (the shell wrapper evals stdout).
	if res.Performed {
		fmt.Fprintf(os.Stderr, "Promoted: %s → %s\n", res.SrcPath, res.DestPath)
		// Emit a cd command on stdout — if the user's wrapper routes
		// `try promote` through eval (see init.go), this auto-cd's them
		// into the new location.
		fmt.Fprintf(os.Stdout, "cd %s\n", shellQuotePath(res.DestPath))
	} else {
		fmt.Fprintf(os.Stderr, "[dry-run] would move: %s → %s\n", res.SrcPath, res.DestPath)
	}
	return nil
}

// shellQuotePath single-quote-wraps a path for safe splicing into a shell
// command-substitution context. Identical scheme to internal/shell's
// shellQuote, duplicated here to avoid exporting an internal helper for one
// caller.
func shellQuotePath(p string) string {
	// Replace any embedded single quote with the standard '\'' sequence.
	out := make([]byte, 0, len(p)+2)
	out = append(out, '\'')
	for i := 0; i < len(p); i++ {
		if p[i] == '\'' {
			out = append(out, '\'', '\\', '\'', '\'')
			continue
		}
		out = append(out, p[i])
	}
	out = append(out, '\'')
	return string(out)
}
