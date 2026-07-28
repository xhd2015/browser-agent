# Scenario

**Feature**: session page client-side install browser detect (UA + content-script marker over chrome boot)

```
# Pure UA helper
Test Client -> read detectRuntimeBrowser in SessionPageApp / ui helpers
  Firefox/ UA -> firefox; else chrome

# Priority resolver
Test Client -> read resolveInstallBrowser
  forced -> __BROWSER_AGENT_EXT__.browser -> detectRuntimeBrowser/UA
  -> snap/boot/path -> default chrome

# SPA wiring
Test Client -> read SessionPageApp + InstallGuideline
  installBrowser = resolveInstallBrowser(...)
  InstallGuideline browser={installBrowser}
  dual chrome + firefox install paths kept
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Tree root is `tests/browser-agent-session-page-browser-detect/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- No real browser; no network; no npm — read-only FS probes.
- Classic TDD: Phase 1 helpers/priority **not** implemented yet → expect **RED**.
- Do not prefer chrome-stamped boot over live UA / `__BROWSER_AGENT_EXT__`.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Leave `Mode` / `ReactProbe` for grouping and leaf Setup.

## Context

- Spec version **0.0.2**.
- Parallel-safe: ModuleRoot reads only.
- Shared helpers below available to all descendant Assert/Setup packages.
- Phase 2 (server Create stamp) is out of scope for this tree.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	return nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport/setup error: %v", err)
	}
}

func assertSourcePresent(t *testing.T, req *Request, resp *Response, label string) string {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.FileExists || strings.TrimSpace(resp.CombinedText) == "" {
		t.Fatalf("%s source missing under ModuleRoot=%s; err=%q found=%v",
			label, req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	// Prefer SessionPageText for helper contracts; fall back to CombinedText.
	if strings.TrimSpace(resp.SessionPageText) != "" {
		return resp.SessionPageText
	}
	return resp.CombinedText
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// sessionSPA returns SessionPageApp (+ helper modules) text without InstallGuideline noise.
func sessionSPA(resp *Response) string {
	if resp == nil {
		return ""
	}
	if strings.TrimSpace(resp.SessionPageText) != "" {
		return resp.SessionPageText
	}
	return resp.CombinedText
}
```
