// Sign builds a Firefox extension package and submits it to AMO for signing
// (unlisted channel). Secrets are read from .firefox-secretes.txt.
//
// Usage (from module root or worktree):
//
//	go run ./script/browser-agent/firefox/sign
//
// Secrets file (gitignored; often kept on the main clone):
//
//	# get from https://addons.mozilla.org/en-US/developers/addon/api/key
//	jwt_issuer=user:12345678:123
//	jwt_secret=...
//
// Search order for .firefox-secretes.txt:
//  1. $BROWSER_AGENT_FIREFOX_SECRETS (explicit path)
//  2. {module-root}/.firefox-secretes.txt
//  3. {git-common-dir parent}/.firefox-secretes.txt  (main repo when in a worktree)
//
// Outputs:
//  - dist/firefox-package/   staged unsigned sources
//  - dist/signed/            AMO-signed .xpi (from web-ext)
//  - browseragent/embedded/firefox-xpi/browser-agent.xpi  //go:embed payload
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/browser-agent/browseragent"
)

const secretsFileName = ".firefox-secretes.txt"

const amoAPIKeyURL = "https://addons.mozilla.org/en-US/developers/addon/api/key/"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
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

	// 1) Secrets
	secretsPath, err := resolveSecretsPath(root)
	if err != nil {
		return err
	}
	if secretsPath == "" {
		printSecretsMissing(root)
		return fmt.Errorf("missing %s — create AMO API credentials and save them, then re-run", secretsFileName)
	}
	fmt.Printf("Using secrets: %s\n", secretsPath)

	issuer, secret, err := readSecrets(secretsPath)
	if err != nil {
		return err
	}
	if issuer == "" || secret == "" {
		printSecretsMissing(root)
		return fmt.Errorf("%s must set jwt_issuer and jwt_secret (see %s)", secretsFileName, amoAPIKeyURL)
	}
	fmt.Println("Secrets: jwt_issuer and jwt_secret present")

	// 2) Generate version stamps
	fmt.Println("==> generate (sync VERSION.txt)")
	if err := runCmd(root, "go", "run", "./script/generate"); err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	// 3) Build extension public → build + bundle-sum
	fmt.Println("==> build Firefox extension shell + bundle-sum")
	buildDir, err := browseragent.BuildFirefoxExtensionShell(root)
	if err != nil {
		return fmt.Errorf("BuildFirefoxExtensionShell: %w", err)
	}
	ver := strings.TrimSpace(browseragent.ClientVersion())
	if _, err := browseragent.EnsureExtensionBundleSum(buildDir, ver); err != nil {
		return fmt.Errorf("EnsureExtensionBundleSum: %w", err)
	}
	fmt.Printf("Staged package: %s (version %s)\n", buildDir, ver)

	// 4) Copy clean package into dist/firefox-package
	stageDir := filepath.Join(root, "dist", "firefox-package")
	if err := os.RemoveAll(stageDir); err != nil {
		return err
	}
	if err := copyDir(buildDir, stageDir); err != nil {
		return fmt.Errorf("stage package: %w", err)
	}
	// Ensure only extension files (no junk)
	fmt.Printf("Unsigned sources: %s\n", stageDir)

	// Also write a local unsigned .xpi for convenience
	unsignedXPI := filepath.Join(root, "dist", fmt.Sprintf("browser-agent-firefox-%s.xpi", ver))
	if err := zipDir(stageDir, unsignedXPI); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not write unsigned xpi: %v\n", err)
	} else {
		fmt.Printf("Unsigned xpi: %s\n", unsignedXPI)
	}

	// 5) web-ext sign → dist/signed
	outDir := filepath.Join(root, "dist", "signed")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	fmt.Println("==> web-ext sign (AMO unlisted channel)")
	if err := runWebExtSign(root, stageDir, outDir, issuer, secret); err != nil {
		return err
	}

	// List results
	entries, _ := os.ReadDir(outDir)
	fmt.Println("==> done")
	fmt.Printf("Signed artifacts in %s:\n", outDir)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, _ := e.Info()
		size := int64(0)
		if info != nil {
			size = info.Size()
		}
		fmt.Printf("  %s  (%d bytes)\n", filepath.Join(outDir, e.Name()), size)
	}

	// Stage into //go:embed so `go install` / install-firefox-extension ship the xpi.
	src, err := browseragent.FindSignedXPIUnder(outDir)
	if err != nil {
		return fmt.Errorf("signed xpi not found after web-ext: %w", err)
	}
	// Prefer a stable name copy under dist/signed for operators.
	stable := filepath.Join(outDir, fmt.Sprintf("browser-agent-%s-signed.xpi", ver))
	if src != stable {
		if data, rerr := os.ReadFile(src); rerr == nil {
			_ = os.WriteFile(stable, data, 0o644)
			src = stable
		}
	}
	embedDir, err := browseragent.StageFirefoxXPIEmbed(root, src, ver)
	if err != nil {
		return fmt.Errorf("stage firefox-xpi embed: %w", err)
	}
	fmt.Printf("Staged for //go:embed: %s/browser-agent.xpi\n", embedDir)
	fmt.Println("Next: go install ./cmd/browser-agent   # or go run ./script/browser-agent/install")
	return nil
}

func printHelp() {
	fmt.Print(`Usage: go run ./script/browser-agent/firefox/sign

Bundle the Firefox extension, sign it with Mozilla AMO (unlisted), and write
artifacts under dist/signed/.

Secrets file: .firefox-secretes.txt  (note spelling)

  # get from https://addons.mozilla.org/en-US/developers/addon/api/key/
  jwt_issuer=user:XXXXXXXX:XXX
  jwt_secret=your-secret

Search order:
  1. $BROWSER_AGENT_FIREFOX_SECRETS
  2. {module}/.firefox-secretes.txt
  3. {main-repo}/.firefox-secretes.txt (when running from a git worktree)

Also writes:
  dist/firefox-package/                         unsigned sources used for signing
  dist/browser-agent-firefox-*.xpi              local unsigned zip (convenience)
  browseragent/embedded/firefox-xpi/browser-agent.xpi  //go:embed payload
`)
}

func printSecretsMissing(root string) {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Firefox AMO signing credentials not found.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Create API credentials:")
	fmt.Fprintln(os.Stderr, "  1. Sign in at https://addons.mozilla.org/developers/")
	fmt.Fprintln(os.Stderr, "  2. Open  "+amoAPIKeyURL)
	fmt.Fprintln(os.Stderr, "  3. Generate a JWT issuer + secret")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Save them as "+secretsFileName+" (gitignored), for example on the main clone:")
	fmt.Fprintln(os.Stderr, "  /path/to/browser-agent/.firefox-secretes.txt")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "File format:")
	fmt.Fprintln(os.Stderr, "  # get from "+amoAPIKeyURL)
	fmt.Fprintln(os.Stderr, "  jwt_issuer=user:12345678:123")
	fmt.Fprintln(os.Stderr, "  jwt_secret=xxxxxxxx")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintf(os.Stderr, "Current module root: %s\n", root)
	fmt.Fprintln(os.Stderr, "Optional: export BROWSER_AGENT_FIREFOX_SECRETS=/absolute/path/to/.firefox-secretes.txt")
	fmt.Fprintln(os.Stderr, "")
}

func resolveSecretsPath(root string) (string, error) {
	if p := strings.TrimSpace(os.Getenv("BROWSER_AGENT_FIREFOX_SECRETS")); p != "" {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
		return "", fmt.Errorf("BROWSER_AGENT_FIREFOX_SECRETS=%s: not a file", p)
	}

	candidates := []string{
		filepath.Join(root, secretsFileName),
	}
	// Git worktree: common dir is main .git; parent of that is main working tree.
	if mainRoot, err := gitMainWorktreeRoot(root); err == nil && mainRoot != "" && mainRoot != root {
		candidates = append(candidates, filepath.Join(mainRoot, secretsFileName))
	}
	// Explicit well-known main clone (fallback).
	candidates = append(candidates, filepath.Join("/Users/xhd2015/Projects/xhd2015/browser-agent", secretsFileName))

	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c, nil
		}
	}
	return "", nil
}

func gitMainWorktreeRoot(moduleRoot string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--path-format=absolute", "--git-common-dir")
	cmd.Dir = moduleRoot
	out, err := cmd.Output()
	if err != nil {
		// older git without --path-format
		cmd = exec.Command("git", "rev-parse", "--git-common-dir")
		cmd.Dir = moduleRoot
		out, err = cmd.Output()
		if err != nil {
			return "", err
		}
	}
	common := strings.TrimSpace(string(out))
	if common == "" {
		return "", fmt.Errorf("empty git-common-dir")
	}
	if !filepath.IsAbs(common) {
		common = filepath.Join(moduleRoot, common)
	}
	// common is .../repo/.git  → main worktree is parent
	if filepath.Base(common) == ".git" {
		return filepath.Dir(common), nil
	}
	// bare or worktree: .../repo/.git  already; if common ends with .git, parent is root
	return filepath.Dir(common), nil
}

func readSecrets(path string) (issuer, secret string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// allow KEY=value or key: value
		var k, v string
		if i := strings.IndexByte(line, '='); i >= 0 {
			k = strings.TrimSpace(line[:i])
			v = strings.TrimSpace(line[i+1:])
		} else if i := strings.IndexByte(line, ':'); i >= 0 {
			k = strings.TrimSpace(line[:i])
			v = strings.TrimSpace(line[i+1:])
		} else {
			continue
		}
		// strip optional quotes
		v = strings.Trim(v, `"'`)
		switch strings.ToLower(k) {
		case "jwt_issuer", "api_key", "api-key", "web_ext_api_key":
			issuer = v
		case "jwt_secret", "api_secret", "api-secret", "web_ext_api_secret":
			secret = v
		}
	}
	if err := sc.Err(); err != nil {
		return "", "", err
	}
	return issuer, secret, nil
}

func runWebExtSign(root, sourceDir, artifactsDir, apiKey, apiSecret string) error {
	// Prefer npx web-ext so node need not install globally.
	args := []string{
		"--yes", "web-ext@8", "sign",
		"--source-dir", sourceDir,
		"--artifacts-dir", artifactsDir,
		"--channel", "unlisted",
		"--api-key", apiKey,
		"--api-secret", apiSecret,
	}
	cmd := exec.Command("npx", args...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Do not leak secrets via verbose env dump; only pass what's needed.
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npx web-ext sign failed: %w\n  hint: check jwt_issuer/jwt_secret at %s", err, amoAPIKeyURL)
	}
	return nil
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

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		// skip junk
		base := filepath.Base(path)
		if base == ".DS_Store" || strings.HasSuffix(base, "~") {
			return nil
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

func zipDir(srcDir, zipPath string) error {
	// Use system zip for simplicity (same as manual packaging).
	absZip, err := filepath.Abs(zipPath)
	if err != nil {
		return err
	}
	_ = os.Remove(absZip)
	cmd := exec.Command("zip", "-r", "-FS", absZip, ".")
	cmd.Dir = srcDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
