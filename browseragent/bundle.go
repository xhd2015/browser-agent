package browseragent

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Default embed destination paths relative to project Root.
const (
	DefaultEmbedExtensionRel = "browseragent/embedded/extension"
	DefaultEmbedSessionRel   = "browseragent/embedded/session-page"
)

// BundleOptions configures staging of extension + session-page embed trees.
type BundleOptions struct {
	Root       string // project root (tests use t.TempDir())
	UseFixture bool   // stage mini fixtures; no npm

	// Optional absolute fixture sources when UseFixture (tests set these).
	// Empty → discover under Root or package defaults.
	FixtureExtensionDir   string
	FixtureSessionPageDir string

	// Dest relative to Root (defaults):
	//   EmbedExtensionRel = "browseragent/embedded/extension"
	//   EmbedSessionRel   = "browseragent/embedded/session-page"
	EmbedExtensionRel string
	EmbedSessionRel   string
}

// BundleResult is the outcome of Bundle staging.
type BundleResult struct {
	ExtensionDir   string // absolute path staged for embed
	SessionPageDir string
	UsedFixture    bool
}

// Bundle stages extension and session-page content into embed paths under Root.
//
// With UseFixture=true: copies mini fixtures (no npm). Idempotent: a second
// call with the same options succeeds and returns the same absolute paths.
func Bundle(opts BundleOptions) (*BundleResult, error) {
	root := strings.TrimSpace(opts.Root)
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("resolve Root: %w", err)
		}
		root = cwd
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve Root: %w", err)
	}

	extRel := strings.TrimSpace(opts.EmbedExtensionRel)
	if extRel == "" {
		extRel = DefaultEmbedExtensionRel
	}
	sessRel := strings.TrimSpace(opts.EmbedSessionRel)
	if sessRel == "" {
		sessRel = DefaultEmbedSessionRel
	}

	extDest := filepath.Join(absRoot, filepath.FromSlash(extRel))
	sessDest := filepath.Join(absRoot, filepath.FromSlash(sessRel))

	usedFixture := opts.UseFixture
	if !opts.UseFixture {
		// Prefer real sources when available; fall back to fixtures.
		staged, ferr := stageRealOrFixture(absRoot, opts, extDest, sessDest)
		if ferr != nil {
			return nil, ferr
		}
		return staged, nil
	}

	extSrc, err := resolveFixtureExtensionDir(absRoot, opts.FixtureExtensionDir)
	if err != nil {
		return nil, err
	}
	sessSrc, err := resolveFixtureSessionPageDir(absRoot, opts.FixtureSessionPageDir)
	if err != nil {
		return nil, err
	}

	if err := stageDir(extSrc, extDest); err != nil {
		return nil, fmt.Errorf("stage extension: %w", err)
	}
	if err := stageDir(sessSrc, sessDest); err != nil {
		return nil, fmt.Errorf("stage session-page: %w", err)
	}
	_ = normalizeSessionPageDist(sessDest)
	_ = ensureCanonicalSessionAssets(sessDest)

	absExt, err := filepath.Abs(extDest)
	if err != nil {
		return nil, err
	}
	// Write/refresh bundle-sum.js for the staged extension package.
	if _, err := EnsureExtensionBundleSum(absExt, ""); err != nil {
		return nil, fmt.Errorf("write extension bundle-sum: %w", err)
	}
	absSess, err := filepath.Abs(sessDest)
	if err != nil {
		return nil, err
	}

	return &BundleResult{
		ExtensionDir:   absExt,
		SessionPageDir: absSess,
		UsedFixture:    usedFixture,
	}, nil
}

func stageRealOrFixture(absRoot string, opts BundleOptions, extDest, sessDest string) (*BundleResult, error) {
	usedFixture := false
	extFromFixture := false
	sessFromFixture := false

	// --- Extension: build public→build (no npm), then stage ---
	extSrc := ""
	if built, err := BuildExtensionShell(absRoot); err == nil {
		extSrc = built
		fmt.Fprintln(os.Stderr, "browser-agent bundle: staged extension from Chrome-Ext-Browser-Agent/public")
	} else {
		// Pre-existing build/ ok if present.
		buildDir := filepath.Join(absRoot, "Chrome-Ext-Browser-Agent", "build")
		if st, err2 := os.Stat(filepath.Join(buildDir, "manifest.json")); err2 == nil && !st.IsDir() {
			extSrc = buildDir
			fmt.Fprintln(os.Stderr, "browser-agent bundle: using existing Chrome-Ext-Browser-Agent/build")
		} else {
			src, rerr := resolveFixtureExtensionDir(absRoot, opts.FixtureExtensionDir)
			if rerr != nil {
				return nil, fmt.Errorf("extension build failed (%v) and no fixture: %w", err, rerr)
			}
			extSrc = src
			extFromFixture = true
			fmt.Fprintf(os.Stderr, "browser-agent bundle: extension build unavailable (%v); staging fixture\n", err)
		}
	}

	// --- Session-page: vite build under react/, then stage ---
	sessSrc := ""
	if dist, err := BuildSessionPage(absRoot); err == nil {
		sessSrc = dist
		// Ensure index.html for embed contract.
		_ = normalizeSessionPageDist(sessSrc)
		fmt.Fprintln(os.Stderr, "browser-agent bundle: staged session-page from react/dist (vite)")
	} else {
		for _, cand := range []string{
			filepath.Join(absRoot, "react", "dist"),
			filepath.Join(absRoot, "react", "dist", "session-page"),
		} {
			if hasSessionIndex(cand) {
				sessSrc = cand
				_ = normalizeSessionPageDist(sessSrc)
				fmt.Fprintln(os.Stderr, "browser-agent bundle: using existing react/dist")
				break
			}
		}
		if sessSrc == "" {
			src, rerr := resolveFixtureSessionPageDir(absRoot, opts.FixtureSessionPageDir)
			if rerr != nil {
				return nil, fmt.Errorf("session-page build failed (%v) and no fixture: %w", err, rerr)
			}
			sessSrc = src
			sessFromFixture = true
			fmt.Fprintf(os.Stderr, "browser-agent bundle: session-page dist unavailable (%v); staging fixture\n", err)
		}
	}

	usedFixture = extFromFixture || sessFromFixture

	if err := stageDir(extSrc, extDest); err != nil {
		return nil, fmt.Errorf("stage extension: %w", err)
	}
	if err := stageDir(sessSrc, sessDest); err != nil {
		return nil, fmt.Errorf("stage session-page: %w", err)
	}
	// Embed contract expects index.html at session-page root.
	_ = normalizeSessionPageDist(sessDest)
	// Canonical /assets/session-page.js for tests + stable URLs (vite uses hashes).
	_ = ensureCanonicalSessionAssets(sessDest)

	absExt, err := filepath.Abs(extDest)
	if err != nil {
		return nil, err
	}
	if _, err := EnsureExtensionBundleSum(absExt, ""); err != nil {
		return nil, fmt.Errorf("write extension bundle-sum: %w", err)
	}
	absSess, err := filepath.Abs(sessDest)
	if err != nil {
		return nil, err
	}
	return &BundleResult{
		ExtensionDir:   absExt,
		SessionPageDir: absSess,
		UsedFixture:    usedFixture,
	}, nil
}

func resolveFixtureExtensionDir(root, override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		p := override
		if !filepath.IsAbs(p) {
			p = filepath.Join(root, p)
		}
		if st, err := os.Stat(filepath.Join(p, "manifest.json")); err == nil && !st.IsDir() {
			return filepath.Abs(p)
		}
		return "", fmt.Errorf("fixture extension missing manifest.json under %s", p)
	}
	candidates := []string{
		// Prefer dedicated fixtures (not the embed dest) so --fixture can re-stage.
		filepath.Join(root, "browseragent", "fixtures", "extension"),
		filepath.Join(root, "tests", "browser-agent-cli-react", "testdata", "mini-extension"),
		filepath.Join(root, "tests", "browser-agent-bundle-dual", "testdata", "mini-extension"),
		filepath.Join(root, "browseragent", "embedded", "extension"),
	}
	for _, c := range candidates {
		if st, err := os.Stat(filepath.Join(c, "manifest.json")); err == nil && !st.IsDir() {
			return filepath.Abs(c)
		}
	}
	return "", fmt.Errorf("no fixture extension found under %s (tried %v)", root, candidates)
}

func resolveFixtureSessionPageDir(root, override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		p := override
		if !filepath.IsAbs(p) {
			p = filepath.Join(root, p)
		}
		if hasSessionIndex(p) {
			return filepath.Abs(p)
		}
		return "", fmt.Errorf("fixture session-page missing index under %s", p)
	}
	candidates := []string{
		filepath.Join(root, "browseragent", "fixtures", "session-page"),
		filepath.Join(root, "tests", "browser-agent-bundle-dual", "testdata", "mini-session-page"),
		filepath.Join(root, "tests", "browser-agent-vite-skill", "testdata", "session-page"),
		filepath.Join(root, "browseragent", "embedded", "session-page"),
	}
	for _, c := range candidates {
		if hasSessionIndex(c) {
			return filepath.Abs(c)
		}
	}
	return "", fmt.Errorf("no fixture session-page found under %s (tried %v)", root, candidates)
}

func hasSessionIndex(dir string) bool {
	for _, name := range []string{"index.html", "session-page.html"} {
		if st, err := os.Stat(filepath.Join(dir, name)); err == nil && !st.IsDir() {
			return true
		}
	}
	return false
}

// stageDir clears dest contents and copies src → dest (recursive).
// If src and dest resolve to the same path, it is a no-op (already staged).
func stageDir(src, dest string) error {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	absDest, err := filepath.Abs(dest)
	if err != nil {
		return err
	}
	if absSrc == absDest {
		// Source is already the embed target (e.g. live module fixture).
		if st, err := os.Stat(absDest); err != nil || !st.IsDir() {
			return fmt.Errorf("stage source/dest %s missing or not a dir: %v", absDest, err)
		}
		return nil
	}
	if err := os.MkdirAll(absDest, 0o755); err != nil {
		return err
	}
	if err := clearDir(absDest); err != nil {
		return err
	}
	return copyTree(absSrc, absDest)
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
