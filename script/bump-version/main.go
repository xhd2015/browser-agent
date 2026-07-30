// Bump root VERSION.txt, then run generate + browser-agent bundle.
//
//	go run ./script/bump-version              # patch +1
//	go run ./script/bump-version --minor
//	go run ./script/bump-version --major
//	go run ./script/bump-version --dry-run
//
// Does not sign Firefox, go install, or create git commits/tags.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

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
	var dryRun, skipBundle, major, minor bool
	for _, a := range args {
		switch a {
		case "-h", "--help":
			fmt.Print(helpText())
			return nil
		case "--dry-run":
			dryRun = true
		case "--skip-bundle":
			skipBundle = true
		case "--major":
			major = true
		case "--minor":
			minor = true
		default:
			if strings.HasPrefix(a, "-") {
				return fmt.Errorf("unknown flag %s (try --help)", a)
			}
			return fmt.Errorf("unexpected argument %q (try --help)", a)
		}
	}
	if major && minor {
		return fmt.Errorf("cannot combine --major and --minor")
	}

	root, err := findModuleRoot()
	if err != nil {
		return err
	}

	verPath := filepath.Join(root, "VERSION.txt")
	oldVer, err := readVersionFile(verPath)
	if err != nil {
		return err
	}
	newVer, err := bumpVersion(oldVer, major, minor)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Printf("%s %s %s %s %s\n",
			c.Gray("would bump"),
			c.Gray("VERSION.txt"),
			c.Gray(oldVer),
			c.Gray("→"),
			c.Green(newVer),
		)
		fmt.Println(c.Gray("(no writes)"))
		return nil
	}

	fmt.Printf("%s %s  %s %s %s\n",
		c.Gray("==>"),
		c.Gray("bump VERSION.txt"),
		c.Gray(oldVer),
		c.Gray("→"),
		c.Green(newVer),
	)
	if err := os.WriteFile(verPath, []byte(newVer+"\n"), 0o644); err != nil {
		return fmt.Errorf("write VERSION.txt: %w", err)
	}

	fmt.Printf("%s %s\n", c.Gray("==>"), c.Gray("generate"))
	if err := runCmd(root, "go", "run", "./script/generate"); err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	if skipBundle {
		fmt.Printf("%s %s\n", c.Gray("==>"), c.Yellow("skip bundle (--skip-bundle)"))
	} else {
		fmt.Printf("%s %s\n", c.Gray("==>"), c.Gray("bundle (extension + session-page)"))
		if err := runCmd(root, "go", "run", "./script/browser-agent/bundle"); err != nil {
			return fmt.Errorf("bundle: %w", err)
		}
	}

	fmt.Printf("%s VERSION.txt is %s", c.Green("ok:"), c.Green(newVer))
	if skipBundle {
		fmt.Printf(" %s\n", c.Gray("(generate done; bundle skipped)"))
	} else {
		fmt.Printf(" %s\n", c.Gray("(generate + bundle done)"))
	}
	fmt.Println(c.Gray("Next:"))
	fmt.Println(c.Gray("  go run ./script/browser-agent/firefox/sign   # if you need AMO-signed xpi"))
	fmt.Println(c.Gray("  go run ./script/browser-agent/install --skip-firefox-sign"))
	return nil
}

func helpText() string {
	return `Usage: go run ./script/bump-version [options]

Bump root VERSION.txt, then run generate and browser-agent bundle.

Default bumps patch (1.0.6 → 1.0.7). Does not sign Firefox, go install, or git commit.

Options:
  (default)               bump patch
  --minor                 bump minor, reset patch (1.0.6 → 1.1.0)
  --major                 bump major, reset minor/patch (1.0.6 → 2.0.0)
  --dry-run               print new version only (no writes)
  --skip-bundle           bump + generate only (no vite/embed)
` + clicolor.FlagHelp + `  -h, --help              Show this help
`
}

func readVersionFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read VERSION.txt: %w", err)
	}
	ver := strings.TrimSpace(string(b))
	ver = strings.TrimPrefix(ver, "v")
	ver = strings.TrimPrefix(ver, "V")
	ver = strings.TrimSpace(ver)
	if ver == "" {
		return "", fmt.Errorf("VERSION.txt is empty")
	}
	return ver, nil
}

// bumpVersion returns major.minor.patch after bump.
// major=true → major+1, minor=0, patch=0
// minor=true → minor+1, patch=0
// else patch+1
func bumpVersion(ver string, major, minor bool) (string, error) {
	if major && minor {
		return "", fmt.Errorf("cannot combine --major and --minor")
	}
	ver = strings.TrimSpace(ver)
	if i := strings.IndexAny(ver, "-+"); i >= 0 {
		return "", fmt.Errorf("pre-release / build metadata not supported: %q", ver)
	}
	parts := strings.Split(ver, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("VERSION.txt must be major.minor.patch (got %q)", ver)
	}
	var nums [3]int
	for i := 0; i < 3; i++ {
		n, err := strconv.Atoi(strings.TrimSpace(parts[i]))
		if err != nil || n < 0 {
			return "", fmt.Errorf("VERSION.txt must be major.minor.patch integers (got %q)", ver)
		}
		nums[i] = n
	}
	switch {
	case major:
		nums[0]++
		nums[1] = 0
		nums[2] = 0
	case minor:
		nums[1]++
		nums[2] = 0
	default:
		nums[2]++
	}
	return fmt.Sprintf("%d.%d.%d", nums[0], nums[1], nums[2]), nil
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
			if _, err := os.Stat(filepath.Join(dir, "VERSION.txt")); err != nil {
				return "", fmt.Errorf("%s: VERSION.txt not found (not module root?)", dir)
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
