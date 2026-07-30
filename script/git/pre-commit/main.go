// Pre-commit helper: ensure go:embed placeholders, sync version sinks, stage both.
//
//	go run ./script/git/pre-commit
//
// Install (managed hook via git-hooks):
//
//	git-hooks pre-commit add 'script.git.pre-commit' go run ./script/git/pre-commit
//
// Steps:
//  1. Create empty placeholder.txt under browseragent/embedded/* when missing
//  2. go run ./script/generate --git-add-generated  (version sinks only)
//  3. git add placeholders
//
// Never git-adds paths outside generate's allowlist + the three placeholders.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/browser-agent/script/internal/clicolor"
)

// Relative to repo root. Keep //go:embed dirs non-empty for clean clones.
var placeholders = []string{
	"browseragent/embedded/extension/placeholder.txt",
	"browseragent/embedded/session-page/placeholder.txt",
	"browseragent/embedded/firefox-xpi/placeholder.txt",
}

func main() {
	mode, remain, err := clicolor.ParseFlags(os.Args[1:])
	c := clicolor.NewStyle(mode)
	if err != nil {
		c.ErrorLine(os.Stderr, err.Error())
		os.Exit(1)
	}
	if err := run(remain, mode, c); err != nil {
		c.ErrorLine(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(args []string, mode clicolor.ColorMode, c clicolor.Style) error {
	for _, a := range args {
		switch a {
		case "-h", "--help":
			fmt.Print(helpText())
			return nil
		default:
			if strings.HasPrefix(a, "-") {
				return fmt.Errorf("unrecognized flag %s (try --help)", a)
			}
			return fmt.Errorf("unexpected argument %q (try --help)", a)
		}
	}

	root, err := gitRoot()
	if err != nil {
		return err
	}

	var toAdd []string
	for _, rel := range placeholders {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := ensurePlaceholder(abs); err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		toAdd = append(toAdd, rel)
	}

	// Stamp VERSION.txt → sinks and stage only generate's allowlisted paths.
	// Forward forced color flags so nested generate matches this process.
	fmt.Printf("%s %s\n", c.Gray("==>"), c.Gray("generate --git-add-generated"))
	genArgs := []string{"run", "./script/generate", "--git-add-generated"}
	switch mode {
	case clicolor.ColorAlways:
		genArgs = append(genArgs, "--color")
	case clicolor.ColorNever:
		genArgs = append(genArgs, "--no-color")
	}
	gen := exec.Command("go", genArgs...)
	gen.Dir = root
	gen.Stdout = os.Stdout
	gen.Stderr = os.Stderr
	if err := gen.Run(); err != nil {
		return fmt.Errorf("go run ./script/generate --git-add-generated: %w", err)
	}

	if err := gitAdd(root, toAdd); err != nil {
		return err
	}
	fmt.Printf("%s %s\n", c.Gray("pre-commit:"), c.Green("placeholders staged"))
	return nil
}

func helpText() string {
	return `Usage: go run ./script/git/pre-commit [options]

Ensure embed placeholders exist, run generate --git-add-generated, and stage
placeholders. Does not git-add paths outside generate sinks + placeholders.

Options:
` + clicolor.FlagHelp + `  -h, --help              Show this help
`
}

func gitRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %w", err)
	}
	root := strings.TrimSpace(string(out))
	if root == "" {
		return "", fmt.Errorf("empty git toplevel")
	}
	return root, nil
}

func ensurePlaceholder(abs string) error {
	if st, err := os.Stat(abs); err == nil {
		if st.IsDir() {
			return fmt.Errorf("is a directory")
		}
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	return os.WriteFile(abs, nil, 0o644)
}

func gitAdd(root string, relPaths []string) error {
	args := append([]string{"add", "--"}, relPaths...)
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	// Quiet: discard stdout; surface stderr only on failure.
	cmd.Stdout = nil
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("git add: %s", msg)
		}
		return fmt.Errorf("git add: %w", err)
	}
	return nil
}
