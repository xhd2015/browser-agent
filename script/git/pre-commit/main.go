// Pre-commit helper: ensure go:embed placeholder files exist and are staged.
//
//	go run ./script/git/pre-commit
//
// Install (managed hook via git-hooks):
//
//	git-hooks pre-commit add 'script.git.pre-commit' go run ./script/git/pre-commit
//
// Silent on success. Creates empty placeholder.txt under browseragent/embedded
// trees when missing, then git-adds them.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Relative to repo root. Keep //go:embed dirs non-empty for clean clones.
var placeholders = []string{
	"browseragent/embedded/extension/placeholder.txt",
	"browseragent/embedded/session-page/placeholder.txt",
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
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

	return gitAdd(root, toAdd)
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
