// Sign submits the Firefox extension to AMO (unlisted channel), or reuses an
// already-signed version from AMO. Full sign builds via firefox/bundle first.
//
// Usage (from module root or worktree):
//
//	go run ./script/browser-agent/firefox/sign
//	go run ./script/browser-agent/firefox/sign --status
//	go run ./script/browser-agent/firefox/sign --force
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
//   - dist/firefox-package/   staged unsigned sources (via firefox/bundle on full sign)
//   - dist/signed/            AMO-signed .xpi (from web-ext or download)
//   - browseragent/embedded/firefox-xpi/browser-agent.xpi  //go:embed payload
//
// Full sign reuses: go run ./script/browser-agent/firefox/bundle
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/browser-agent/browseragent"
	"github.com/xhd2015/browser-agent/script/internal/clicolor"
)

const secretsFileName = ".firefox-secretes.txt"

const amoAPIKeyURL = "https://addons.mozilla.org/en-US/developers/addon/api/key/"

// defaultGeckoID matches Firefox-Ext-Browser-Agent/public/manifest.json when present.
const defaultGeckoID = "browser-agent@xhd2015"

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
	var statusOnly, force bool
	for _, a := range args {
		switch a {
		case "-h", "--help":
			printHelp()
			return nil
		case "--status":
			statusOnly = true
		case "--force":
			force = true
		default:
			if strings.HasPrefix(a, "-") {
				return fmt.Errorf("unknown flag %s (try --help)", a)
			}
			return fmt.Errorf("unexpected argument %q (try --help)", a)
		}
	}
	if statusOnly && force {
		return fmt.Errorf("cannot combine --status and --force")
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
	fmt.Printf("%s %s\n", c.Gray("Using secrets:"), secretsPath)

	issuer, secret, err := readSecrets(secretsPath)
	if err != nil {
		return err
	}
	if issuer == "" || secret == "" {
		printSecretsMissing(root)
		return fmt.Errorf("%s must set jwt_issuer and jwt_secret (see %s)", secretsFileName, amoAPIKeyURL)
	}
	fmt.Printf("%s %s\n", c.Gray("Secrets:"), c.Green("jwt_issuer and jwt_secret present"))

	wantVer := browseragent.ReadProductVersion(root)
	if wantVer == "" {
		// Light sync so VERSION sinks exist for full sign; status can still use file.
		_ = runCmd(root, "go", "run", "./script/generate")
		wantVer = browseragent.ReadProductVersion(root)
	}
	if wantVer == "" {
		wantVer = strings.TrimSpace(browseragent.ClientVersion())
	}
	if wantVer == "" {
		return fmt.Errorf("product version empty (set VERSION.txt)")
	}
	guid := readGeckoID(root)
	fmt.Printf("%s %s\n", c.Gray("product VERSION.txt:"), c.Green(wantVer))
	fmt.Printf("%s %s\n", c.Gray("addon:"), c.Gray(guid))

	// 2) Query AMO (status mode or smart reuse)
	fmt.Printf("%s %s %s\n", c.Gray("==>"), c.Gray("AMO lookup"), c.Gray(wantVer))
	st, err := lookupAMOVersion(guid, wantVer, issuer, secret)
	if err != nil {
		return err
	}

	if statusOnly {
		return printStatus(c, st)
	}

	if !force && st.State == "signed" && st.FileURL != "" {
		return reuseSignedFromAMO(root, wantVer, st, issuer, secret, c)
	}
	if !force && st.State == "pending" {
		return fmt.Errorf("AMO version %s is still pending (validation/approval)\n  re-run: go run ./script/browser-agent/firefox/sign --status\n  or wait and re-run sign to download when ready (AMO can take hours)", wantVer)
	}
	if force {
		fmt.Printf("%s %s\n", c.Gray("==>"), c.Yellow("force full sign (--force)"))
	} else if st.State == "not_found" {
		fmt.Printf("%s %s\n", c.Gray("==>"), c.Gray("AMO has no signed "+wantVer+" — building and submitting"))
	} else {
		fmt.Printf("%s %s (%s)\n", c.Gray("==>"), c.Gray("AMO not reusable"), st.State)
	}

	return fullSign(root, wantVer, issuer, secret, c)
}

func printStatus(c clicolor.Style, st *AMOVersionStatus) error {
	fmt.Printf("%s %s\n", c.Gray("AMO version"), c.Green(st.Version))
	switch st.State {
	case "signed":
		fmt.Printf("  state: %s\n", c.Green("signed"))
		if st.FileURL != "" {
			fmt.Printf("  file:  %s\n", c.Gray(st.FileURL))
		}
		if st.Channel != "" {
			fmt.Printf("  channel: %s\n", c.Gray(st.Channel))
		}
		fmt.Println(c.Gray("  hint: go run ./script/browser-agent/firefox/sign  # download + stage without rebuild"))
		return nil
	case "pending":
		fmt.Printf("  state: %s\n", c.Yellow("pending"))
		if st.FileStatus != "" {
			fmt.Printf("  file_status: %s\n", c.Gray(st.FileStatus))
		}
		fmt.Println(c.Gray("  hint: wait for AMO (can take hours), then re-run --status or sign to download"))
		return fmt.Errorf("version %s still pending on AMO", st.Version)
	case "not_found":
		fmt.Printf("  state: %s\n", c.Yellow("not found"))
		fmt.Println(c.Gray("  hint: go run ./script/browser-agent/firefox/sign  # full build + web-ext sign"))
		return fmt.Errorf("version %s not found on AMO", st.Version)
	case "rejected":
		fmt.Printf("  state: %s\n", c.Red("rejected/disabled"))
		return fmt.Errorf("version %s rejected or disabled on AMO", st.Version)
	default:
		fmt.Printf("  state: %s\n", c.Yellow(st.State))
		return fmt.Errorf("version %s state=%s", st.Version, st.State)
	}
}

func reuseSignedFromAMO(root, wantVer string, st *AMOVersionStatus, issuer, secret string, c clicolor.Style) error {
	fmt.Printf("%s %s\n", c.Gray("==>"), c.Green("AMO already has signed "+wantVer+" — downloading"))
	outDir := filepath.Join(root, "dist", "signed")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	stable := filepath.Join(outDir, fmt.Sprintf("browser-agent-%s-signed.xpi", wantVer))
	if err := downloadAMOFile(st.FileURL, stable, issuer, secret); err != nil {
		return fmt.Errorf("download signed xpi: %w", err)
	}
	fmt.Printf("  wrote %s\n", stable)
	if err := browseragent.ValidateReleaseFirefoxXPI(stable, wantVer); err != nil {
		return fmt.Errorf("downloaded xpi failed validation: %w", err)
	}
	embedDir, err := browseragent.StageFirefoxXPIEmbed(root, stable, wantVer)
	if err != nil {
		return fmt.Errorf("stage firefox-xpi embed: %w", err)
	}
	fmt.Printf("  staged %s/browser-agent.xpi\n", embedDir)
	fmt.Printf("%s %s\n", c.Green("ok:"), "reused AMO signed xpi (skipped build + web-ext sign)")
	fmt.Println("Next: go install ./cmd/browser-agent   # or go run ./script/browser-agent/install --skip-firefox-sign")
	return nil
}

func fullSign(root, wantVer, issuer, secret string, c clicolor.Style) error {
	// 2–4) Unsigned package (generate + shell + dist/firefox-package + unsigned xpi)
	fmt.Printf("%s %s\n", c.Gray("==>"), c.Gray("firefox/bundle (unsigned package)"))
	if err := runCmd(root, "go", "run", "./script/browser-agent/firefox/bundle"); err != nil {
		return fmt.Errorf("firefox/bundle: %w", err)
	}
	// Prefer disk VERSION after bundle/generate
	if v := browseragent.ReadProductVersion(root); v != "" {
		wantVer = v
	}
	stageDir := filepath.Join(root, "dist", "firefox-package")
	if st, err := os.Stat(stageDir); err != nil || !st.IsDir() {
		return fmt.Errorf("firefox/bundle did not produce %s", stageDir)
	}

	// 5) web-ext sign → dist/signed
	outDir := filepath.Join(root, "dist", "signed")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	fmt.Printf("%s %s\n", c.Gray("==>"), c.Gray("web-ext sign (AMO unlisted channel)"))
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

	// Stage into //go:embed — pick the xpi that matches wantVer (not an older *signed* leftover).
	src, err := browseragent.FindSignedXPIMatchingVersion(outDir, wantVer)
	if err != nil {
		return fmt.Errorf("signed xpi not found after web-ext for version %s: %w", wantVer, err)
	}
	// Stable operator name for this version only (copy matching bytes; never clobber with another version).
	stable := filepath.Join(outDir, fmt.Sprintf("browser-agent-%s-signed.xpi", wantVer))
	if src != stable {
		data, rerr := os.ReadFile(src)
		if rerr != nil {
			return fmt.Errorf("read signed xpi: %w", rerr)
		}
		if err := os.WriteFile(stable, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", stable, err)
		}
		src = stable
	}
	if err := browseragent.ValidateReleaseFirefoxXPI(src, wantVer); err != nil {
		return fmt.Errorf("signed xpi validation: %w", err)
	}
	embedDir, err := browseragent.StageFirefoxXPIEmbed(root, src, wantVer)
	if err != nil {
		return fmt.Errorf("stage firefox-xpi embed: %w", err)
	}
	fmt.Printf("Staged for //go:embed: %s/browser-agent.xpi\n", embedDir)
	fmt.Println("Next: go install ./cmd/browser-agent   # or go run ./script/browser-agent/install --skip-firefox-sign")
	return nil
}

func readGeckoID(root string) string {
	path := filepath.Join(root, "Firefox-Ext-Browser-Agent", "public", "manifest.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return defaultGeckoID
	}
	var m struct {
		BrowserSpecificSettings struct {
			Gecko struct {
				ID string `json:"id"`
			} `json:"gecko"`
		} `json:"browser_specific_settings"`
	}
	if json.Unmarshal(raw, &m) != nil {
		return defaultGeckoID
	}
	if id := strings.TrimSpace(m.BrowserSpecificSettings.Gecko.ID); id != "" {
		return id
	}
	return defaultGeckoID
}

func printHelp() {
	fmt.Print(`Usage: go run ./script/browser-agent/firefox/sign [options]

Sign the Firefox extension with Mozilla AMO (unlisted), or reuse an already
signed version from AMO for the current VERSION.txt.

Default (smart):
  1. Query AMO for browser-agent@xhd2015 + VERSION.txt
  2. If already signed → download to dist/signed/ + stage embed (skip build)
  3. Else: go run ./script/browser-agent/firefox/bundle, then web-ext sign, stage

Options:
  --status     Query AMO only (no build/sign/stage); exit non-zero if not signed
  --force      Always bundle + web-ext sign (do not reuse AMO download)
` + clicolor.FlagHelp + `  -h, --help   Show this help

Secrets file: .firefox-secretes.txt  (note spelling)

  # get from https://addons.mozilla.org/en-US/developers/addon/api/key/
  jwt_issuer=user:XXXXXXXX:XXX
  jwt_secret=your-secret

Search order:
  1. $BROWSER_AGENT_FIREFOX_SECRETS
  2. {module}/.firefox-secretes.txt
  3. {main-repo}/.firefox-secretes.txt (when running from a git worktree)

Unsigned package only (no AMO):
  go run ./script/browser-agent/firefox/bundle

Also writes (full sign path, after bundle):
  dist/firefox-package/                         unsigned sources used for signing
  dist/browser-agent-firefox-*.xpi              local unsigned zip (from bundle)
  dist/signed/browser-agent-*-signed.xpi        AMO-signed package
  browseragent/embedded/firefox-xpi/browser-agent.xpi  //go:embed payload

Note: AMO signing/audit may take hours. Aborting the client does not cancel AMO;
  re-run --status or sign later to download when ready.
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
		return fmt.Errorf("npx web-ext sign failed: %w\n  hint: check jwt_issuer/jwt_secret and system clock at %s\n  if you aborted earlier, try: go run ./script/browser-agent/firefox/sign --status", err, amoAPIKeyURL)
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
