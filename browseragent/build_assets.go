package browseragent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BuildExtensionShell copies Chrome-Ext-Browser-Agent/public → build/.
// No npm required when public/ exists (shell is static MV3 files).
// Returns absolute path to build/ on success.
func BuildExtensionShell(root string) (buildDir string, err error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	publicDir := filepath.Join(absRoot, "Chrome-Ext-Browser-Agent", "public")
	buildDir = filepath.Join(absRoot, "Chrome-Ext-Browser-Agent", "build")
	if st, err := os.Stat(filepath.Join(publicDir, "manifest.json")); err != nil || st.IsDir() {
		return "", fmt.Errorf("extension public/manifest.json missing under %s", publicDir)
	}
	if err := stageDir(publicDir, buildDir); err != nil {
		return "", fmt.Errorf("stage extension public→build: %w", err)
	}
	return filepath.Abs(buildDir)
}

// BuildSessionPage runs package manager install (if needed) and vite build under react/.
// Returns absolute path to react/dist when session HTML is present.
func BuildSessionPage(root string) (distDir string, err error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	reactDir := filepath.Join(absRoot, "react")
	if st, err := os.Stat(filepath.Join(reactDir, "package.json")); err != nil || st.IsDir() {
		return "", fmt.Errorf("react/package.json missing under %s", reactDir)
	}

	if err := ensureNodeModules(reactDir); err != nil {
		return "", err
	}
	if err := runInDir(reactDir, "npm", "run", "build"); err != nil {
		// Prefer pnpm if npm script failed and pnpm is available.
		if _, lookErr := exec.LookPath("pnpm"); lookErr == nil {
			if err2 := runInDir(reactDir, "pnpm", "run", "build"); err2 != nil {
				return "", fmt.Errorf("vite build failed (npm: %v; pnpm: %w)", err, err2)
			}
		} else {
			return "", fmt.Errorf("vite build failed: %w", err)
		}
	}

	distDir = filepath.Join(reactDir, "dist")
	// Vite multi-page may only emit session-page.html — promote to index.html.
	_ = normalizeSessionPageDist(distDir)
	if !hasSessionIndex(distDir) {
		return "", fmt.Errorf("react/dist missing index.html or session-page.html after build")
	}
	return filepath.Abs(distDir)
}

// normalizeSessionPageDist ensures index.html exists (copy from session-page.html).
func normalizeSessionPageDist(distDir string) error {
	index := filepath.Join(distDir, "index.html")
	if st, err := os.Stat(index); err == nil && !st.IsDir() {
		return nil
	}
	for _, name := range []string{"session-page.html", "session-page/index.html"} {
		src := filepath.Join(distDir, filepath.FromSlash(name))
		if st, err := os.Stat(src); err == nil && !st.IsDir() {
			return copyFile(src, index)
		}
	}
	return fmt.Errorf("no session-page.html to promote to index.html in %s", distDir)
}

// ensureCanonicalSessionAssets writes assets/session-page.js when missing so
// GET /assets/session-page.js stays stable (vite emits hashed names).
func ensureCanonicalSessionAssets(sessionPageDir string) error {
	canonical := filepath.Join(sessionPageDir, "assets", "session-page.js")
	if st, err := os.Stat(canonical); err == nil && !st.IsDir() {
		return nil
	}
	// Prefer copying a vite session-page-*.js chunk when present.
	assetsDir := filepath.Join(sessionPageDir, "assets")
	if entries, err := os.ReadDir(assetsDir); err == nil {
		for _, e := range entries {
			name := e.Name()
			if strings.HasPrefix(name, "session-page-") && strings.HasSuffix(name, ".js") {
				return copyFile(filepath.Join(assetsDir, name), canonical)
			}
		}
	}
	// Minimal poller shim (fixture contract).
	const shim = `// browser-agent session-page asset (canonical path)
(function () {
  function boot() {
    try {
      var el = document.getElementById("browser-agent-boot");
      if (!el) return null;
      return JSON.parse(el.textContent || "{}");
    } catch (e) {
      return null;
    }
  }
  function poll() {
    var b = boot() || window.__BROWSER_AGENT || {};
    var sid = b.session_id || b.sessionId || "";
    var q = sid ? "?session=" + encodeURIComponent(sid) : "";
    fetch("/v1/session" + q).catch(function () {});
  }
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", poll);
  } else {
    poll();
  }
  setInterval(poll, 1000);
})();
`
	if err := os.MkdirAll(filepath.Dir(canonical), 0o755); err != nil {
		return err
	}
	return os.WriteFile(canonical, []byte(shim), 0o644)
}

func ensureNodeModules(reactDir string) error {
	if st, err := os.Stat(filepath.Join(reactDir, "node_modules")); err == nil && st.IsDir() {
		return nil
	}
	// Prefer npm ci/install; fall back to pnpm.
	if err := runInDir(reactDir, "npm", "install"); err == nil {
		return nil
	} else if _, lookErr := exec.LookPath("pnpm"); lookErr == nil {
		return runInDir(reactDir, "pnpm", "install")
	} else {
		return fmt.Errorf("npm install failed and pnpm not found: %w", err)
	}
}

func runInDir(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}
