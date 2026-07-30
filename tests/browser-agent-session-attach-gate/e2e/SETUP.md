# Scenario

**Feature**: Real-browser E2E — session attach gate via playwright-debug

```
RunDaemon(127.0.0.1:0, BaseDir) -> POST /v1/sessions (no Chrome)
ExtractEmbeddedExtension -> playwright-debug --extension --headed
Control tab lifecycle -> eval job success/fail asserts (JSON stdout lines)
```

## Preconditions

- Mode is `e2e`.
- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- `playwright-debug` on PATH; skip leaf when absent.
- Default `ReadyTimeout = 10s`, `PlaywrightTimeout = 90s`.

## Steps

1. Set `Mode = ModeE2E`.
2. Allocate temp `BaseDir` per leaf.
3. Leaf sets `PlaywrightOp` and `SessionID`.

## Context

- Reuses RunDaemon + ExtractEmbeddedExtension harness from
  `browser-agent-e2e-playwright` / `browser-agent-active-tab-routing` /
  `browser-agent-session-tab-targeting`.
- Playwright scripts print JSON assert lines to stdout.
- ASSERT frontmatter: `e2e, slow, ui-automation`.
- Behavioral asserts only (job ok / fail) — no Chrome debugger infobar scraping.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeE2E
	dir := t.TempDir()
	req.BaseDir = filepath.Join(dir, "browser-agent-session-attach-gate-e2e")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 10 * time.Second
	}
	if req.PlaywrightTimeout == 0 {
		req.PlaywrightTimeout = 90 * time.Second
	}
	return nil
}
```
