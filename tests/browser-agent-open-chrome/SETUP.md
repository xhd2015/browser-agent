# Scenario

**Feature**: browser-agent managed Chrome (open-chrome + session new integration)

```
# Layout + extension sync + argv builder
Test Client -> DefaultManagedChromeLayout | LayoutFromRoot(root)
Test Client -> EnsureManagedExtension(layout) -> extensions/{version}/
Test Client -> BuildManagedChromeArgs(dataDir, extPath, url)

# Open + CLI
Test Client -> OpenManagedChrome({LaunchFn}) -> record argv (no real Chrome)
Operator -> HandleCLI(open-chrome [--root] [url]) -> pretty stdout + LaunchFn

# Session new integration
SessionNew -> OpenManagedChrome({URL}) -> ManagedChromeTestHooks.LaunchFn
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Tree root is `tests/browser-agent-open-chrome/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Package `browseragent` exports managed chrome APIs (TDD red until implemented):
  `DefaultManagedChromeLayout`, `LayoutFromRoot`, `EnsureManagedExtension`,
  `BuildManagedChromeArgs`, `OpenManagedChrome`, `OpenChromeResult`.
- `browseragent/inject` exports `ManagedChromeHooks` + `ManagedChromeTestHooks`.
- Each leaf uses isolated temp dirs for custom `--root`; default-root uses real home.
- No real Chrome; `LaunchFn` always injected in open/session leaves.
- `install-chrome-extension` command remains unchanged (not under test here).

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate unique temp `ManagedRoot` / `BaseDir` per leaf when needed.
3. Default `SessionID` for session-new integration.
4. Default `ReadyTimeout` to 5s.
5. Leave `Mode` and op-specific fields for grouping/leaf Setup.

## Context

- Parallel-safe: temp dirs per leaf; session-new starts ephemeral daemon on `:0`.
- Shared assertion helpers below available to all descendant Assert packages.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/browser-agent/browseragent"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	dir := t.TempDir()
	req.ManagedRoot = filepath.Join(dir, "managed-chrome")
	req.BaseDir = filepath.Join(dir, "browser-agent-sessions")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	if req.SessionID == "" {
		req.SessionID = "sess-open-chrome-1"
	}
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	return nil
}

func assertManagedChromeArgsContract(t *testing.T, args []string, dataDir, extPath, url string) {
	t.Helper()
	if len(args) == 0 {
		t.Fatal("ChromeArgs is empty")
	}
	hasUserData := false
	hasLoad := false
	hasNewWindow := false
	for i, a := range args {
		if a == "--user-data-dir" && i+1 < len(args) && args[i+1] == dataDir {
			hasUserData = true
		}
		if strings.HasPrefix(a, "--user-data-dir=") {
			val := strings.TrimPrefix(a, "--user-data-dir=")
			if val == dataDir || strings.Contains(val, dataDir) {
				hasUserData = true
			}
		}
		if a == "--load-extension" && i+1 < len(args) && args[i+1] == extPath {
			hasLoad = true
		}
		if strings.HasPrefix(a, "--load-extension=") {
			val := strings.TrimPrefix(a, "--load-extension=")
			if val == extPath || strings.Contains(a, extPath) {
				hasLoad = true
			}
		}
		if a == "--new-window" {
			hasNewWindow = true
		}
	}
	if !hasUserData {
		t.Fatalf("args missing --user-data-dir=%q; args=%v", dataDir, args)
	}
	if !hasLoad {
		t.Fatalf("args missing --load-extension=%q; args=%v", extPath, args)
	}
	if !hasNewWindow {
		t.Fatalf("args missing --new-window; args=%v", args)
	}
	if url != "" {
		found := false
		for _, a := range args {
			if a == url || strings.Contains(a, url) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("args should include url %q; args=%v", url, args)
		}
	} else {
		for _, a := range args {
			if strings.HasPrefix(a, "http://") || strings.HasPrefix(a, "https://") {
				t.Fatalf("blank window must not include url arg; found %q in %v", a, args)
			}
		}
	}
}

func assertArgsContainUserDataDir(t *testing.T, args []string) {
	t.Helper()
	for i, a := range args {
		if a == "--user-data-dir" || strings.HasPrefix(a, "--user-data-dir=") {
			return
		}
		if a == "--user-data-dir" && i+1 < len(args) {
			return
		}
	}
	t.Fatalf("args missing --user-data-dir; args=%v", args)
}

func assertArgsNoManagedChrome(t *testing.T, args []string) {
	t.Helper()
	for _, a := range args {
		if a == "--user-data-dir" || strings.HasPrefix(a, "--user-data-dir=") {
			t.Fatalf("system chrome must not include --user-data-dir; args=%v", args)
		}
		if a == "--load-extension" || strings.HasPrefix(a, "--load-extension=") {
			t.Fatalf("system chrome must not include --load-extension; args=%v", args)
		}
	}
}

func assertLayoutUnderRoot(t *testing.T, layout browseragent.ManagedChromeLayout, root string) {
	t.Helper()
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	if layout.Root != absRoot {
		t.Fatalf("Layout.Root = %q, want %q", layout.Root, absRoot)
	}
	wantData := filepath.Join(absRoot, "data")
	if layout.DataDir != wantData {
		t.Fatalf("Layout.DataDir = %q, want %q", layout.DataDir, wantData)
	}
	wantExt := filepath.Join(absRoot, "extensions")
	if layout.ExtensionsDir != wantExt {
		t.Fatalf("Layout.ExtensionsDir = %q, want %q", layout.ExtensionsDir, wantExt)
	}
}

func assertContainsFold(t *testing.T, haystack string, needles ...string) {
	t.Helper()
	low := strings.ToLower(haystack)
	for _, n := range needles {
		if !strings.Contains(low, strings.ToLower(n)) {
			t.Fatalf("expected text to contain %q; got:\n%s", n, truncate(haystack, 800))
		}
	}
}

func assertStdoutTrailingNewline(t *testing.T, stdout string) {
	t.Helper()
	if stdout == "" {
		t.Fatal("stdout is empty")
	}
	if !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("stdout must end with \\n; last bytes=%q", tail(stdout, 40))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
```
