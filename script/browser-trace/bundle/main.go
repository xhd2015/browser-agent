// Bundle prepares browsertrace/embedded/extension/** for //go:embed.
//
// Production: builds Chrome-Ext-Capture-API and stages the unpacked build tree.
// Fallback: if the extension build is unavailable, stages the mini fixture from
// tests/browser-trace-embed-extension/testdata/mini-extension/ so embed still compiles.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/xgo/support/cmd"
)

const (
	extensionDir   = "Chrome-Ext-Capture-API"
	embedTargetRel = "browsertrace/embedded/extension"
	miniFixtureRel = "tests/browser-trace-embed-extension/testdata/mini-extension"
)

func main() {
	if err := handle(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	useFixture := false
	for _, a := range args {
		if a == "--fixture" || a == "--mini" {
			useFixture = true
		}
	}

	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}

	embedDir := filepath.Join(root, embedTargetRel)
	if err := os.MkdirAll(embedDir, 0o755); err != nil {
		return err
	}

	if useFixture {
		return stageFixture(root, embedDir)
	}

	// Prefer real extension build when possible.
	extRoot := filepath.Join(root, extensionDir)
	if _, err := os.Stat(extRoot); err != nil {
		fmt.Fprintln(os.Stderr, "Chrome-Ext-Capture-API not found; staging mini fixture")
		return stageFixture(root, embedDir)
	}

	// Build via existing script when npm is available.
	if err := buildExtension(root, extRoot); err != nil {
		fmt.Fprintf(os.Stderr, "extension build failed (%v); staging mini fixture\n", err)
		return stageFixture(root, embedDir)
	}

	buildDir := filepath.Join(extRoot, "build")
	if st, err := os.Stat(buildDir); err != nil || !st.IsDir() {
		fmt.Fprintln(os.Stderr, "extension build/ missing; staging mini fixture")
		return stageFixture(root, embedDir)
	}

	if err := clearDir(embedDir); err != nil {
		return err
	}
	if err := copyTree(buildDir, embedDir); err != nil {
		return fmt.Errorf("stage extension build: %w", err)
	}
	fmt.Printf("Staged Chrome-Ext-Capture-API build into %s\n", embedDir)
	return nil
}

func buildExtension(root, extRoot string) error {
	// Reuse the dedicated build script when present.
	buildMain := filepath.Join(root, "script", "chrome-ext-capture-api", "build", "main.go")
	if _, err := os.Stat(buildMain); err == nil {
		return cmd.Debug().Dir(root).Run("go", "run", "./script/chrome-ext-capture-api/build")
	}
	// Fallback: npm run build in extension dir.
	if _, err := os.Stat(filepath.Join(extRoot, "node_modules")); err != nil {
		if err := cmd.Debug().Dir(extRoot).Run("npm", "install"); err != nil {
			return err
		}
	}
	return cmd.Debug().Dir(extRoot).Run("npm", "run", "build")
}

func stageFixture(root, embedDir string) error {
	src := filepath.Join(root, miniFixtureRel)
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("mini fixture not found at %s: %w", src, err)
	}
	if err := clearDir(embedDir); err != nil {
		return err
	}
	if err := copyTree(src, embedDir); err != nil {
		return fmt.Errorf("stage mini fixture: %w", err)
	}
	fmt.Printf("Staged mini fixture into %s\n", embedDir)
	return nil
}

func clearDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(dir, 0o755)
		}
		return err
	}
	for _, e := range entries {
		// Keep the directory; remove contents so go:embed still has a target.
		if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		// Skip obvious build noise.
		base := info.Name()
		if strings.HasPrefix(base, ".") && base != ".gitkeep" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
