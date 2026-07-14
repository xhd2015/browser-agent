# Scenario

**Feature**: Real-browser E2E via playwright-debug + embedded extension

```
Test Client -> RunDaemon(127.0.0.1:0, BaseDir) -> POST /v1/sessions (no Chrome)
Test Client -> ExtractEmbeddedExtension(BaseDir) -> ExtensionDir
Test Client -> playwright-debug --extension --headless run testdata/*.js
Extension SW -> WS daemon per session tab on /go?session=id
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- Daemon phases 1–9 (`RunDaemon`, `POST /v1/sessions`, per-session extension WS).
- `playwright-debug` on PATH; Chromium installed for Playwright.
- Tree root is `tests/browser-agent-e2e-playwright/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Leaves skip when `playwright-debug` absent (`t.Skip`).
- Headless extension load may fail on some hosts — prefer `--headless` first per spec.
- Classic TDD: leaves may be **RED** until harness + extension wiring are complete.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate temp `BaseDir` for every leaf.
3. Default `ReadyTimeout = 10s`, `PlaywrightTimeout = 60s`.
4. Leaf Setup sets `PlaywrightOp` and session id(s).

## Context

- Spec version **0.0.2**.
- No `openChrome` / `SessionNew` — sessions created only via `POST /v1/sessions`.
- Playwright scripts print JSON assert lines to stdout for harness parsing.
- ASSERT frontmatter labels: `slow, ui-automation` (run via `--label 'slow && ui-automation'`).

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	dir := t.TempDir()
	req.BaseDir = filepath.Join(dir, "browser-agent-e2e")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 10 * time.Second
	}
	if req.PlaywrightTimeout == 0 {
		req.PlaywrightTimeout = 60 * time.Second
	}
	return nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport/setup error: %v", err)
	}
}
```