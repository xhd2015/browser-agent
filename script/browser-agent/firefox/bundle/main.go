// Bundle builds an unsigned Firefox extension package (no AMO sign).
//
// Usage (from module root or worktree):
//
//	go run ./script/browser-agent/firefox/bundle
//
// Outputs:
//   - dist/firefox-package/                    staged unsigned sources
//   - dist/browser-agent-firefox-<VERSION>.xpi  convenience unsigned zip
//
// Build steps are shared with browseragent.Bundle → StageFirefoxExtensionEmbed
// via PrepareFirefoxExtensionBuild (public→build + bundle-sum).
//
// Signing reuses this package:
//
//	go run ./script/browser-agent/firefox/sign
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/browser-agent/browseragent"
	"github.com/xhd2015/browser-agent/script/internal/clicolor"
)

func main() {
	mode, remain, err := clicolor.ParseFlags(os.Args[1:])
	c := clicolor.NewStyle(mode)
	if err != nil {
		c.ErrorLine(os.Stderr, err.Error())
		os.Exit(1)
	}
	if err := run(remain, c); err != nil {
		c.ErrorLine(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(args []string, c clicolor.Style) error {
	for _, a := range args {
		switch a {
		case "-h", "--help":
			printHelp()
			return nil
		default:
			if strings.HasPrefix(a, "-") {
				return fmt.Errorf("unknown flag %s (try --help)", a)
			}
			return fmt.Errorf("unexpected argument %q (try --help)", a)
		}
	}

	root, err := findModuleRoot()
	if err != nil {
		return err
	}

	fmt.Printf("%s %s\n", c.Gray("==>"), c.Gray("generate (sync VERSION.txt)"))
	if err := runCmd(root, "go", "run", "./script/generate"); err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	fmt.Printf("%s %s\n", c.Gray("==>"), c.Gray("PrepareFirefoxExtensionBuild + dist/ package"))
	pkg, err := browseragent.BundleFirefoxUnsignedPackage(root)
	if err != nil {
		return err
	}
	fmt.Printf("Staged package: %s (version %s)\n", pkg.BuildDir, pkg.Version)
	fmt.Printf("Unsigned sources: %s\n", pkg.StageDir)
	if pkg.UnsignedXPI != "" {
		fmt.Printf("Unsigned xpi: %s\n", pkg.UnsignedXPI)
	}

	fmt.Printf("%s %s\n", c.Green("ok:"), "firefox unsigned package ready")
	fmt.Printf("%s %s\n", c.Gray("sources:"), pkg.StageDir)
	if pkg.UnsignedXPI != "" {
		fmt.Printf("%s %s\n", c.Gray("unsigned xpi:"), pkg.UnsignedXPI)
	}
	fmt.Println(c.Gray("Next: go run ./script/browser-agent/firefox/sign   # AMO sign (reuses this package)"))
	return nil
}

func printHelp() {
	fmt.Print(`Usage: go run ./script/browser-agent/firefox/bundle [options]

Build an unsigned Firefox extension package (no AMO credentials, no sign).

Steps:
  1. go run ./script/generate
  2. PrepareFirefoxExtensionBuild (public → build + bundle-sum)
  3. dist/firefox-package/
  4. dist/browser-agent-firefox-<VERSION>.xpi

Shared with browser-agent/bundle (StageFirefoxExtensionEmbed uses the same prepare).

Outputs:
  dist/firefox-package/                         unsigned sources (web-ext input)
  dist/browser-agent-firefox-*.xpi              local unsigned zip

To AMO-sign (reuses this package):
  go run ./script/browser-agent/firefox/sign

` + clicolor.FlagHelp + `  -h, --help   Show this help
`)
}

func runCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func findModuleRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "cmd", "browser-agent")); err != nil {
				return "", fmt.Errorf("%s: cmd/browser-agent not found", dir)
			}
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", wd)
		}
		dir = parent
	}
}
