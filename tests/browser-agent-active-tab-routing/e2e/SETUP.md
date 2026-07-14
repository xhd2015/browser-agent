# Scenario

**Feature**: Real-browser E2E — eval runs on active user tab in session window

```
RunDaemon(127.0.0.1:0, BaseDir) -> POST /v1/sessions (no Chrome)
ExtractEmbeddedExtension -> playwright-debug --extension --headed
Tab1 /go?session=S; Tab2 user URL active -> info+eval jobs target user tab
```

## Preconditions

- Mode is `e2e`.
- Package `github.com/xhd2015/browser-agent/browseragent` importable.
- `playwright-debug` on PATH; skip leaf when absent.
- Default `ReadyTimeout = 10s`, `PlaywrightTimeout = 60s`.

## Steps

1. Set `Mode = ModeE2E`.
2. Allocate temp `BaseDir` per leaf.
3. Leaf sets `PlaywrightOp` and `SessionID`.

## Context

- Reuses RunDaemon + ExtractEmbeddedExtension harness from `browser-agent-e2e-playwright`.
- Playwright scripts print JSON assert lines to stdout.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeE2E
	dir := t.TempDir()
	req.BaseDir = filepath.Join(dir, "browser-agent-active-tab-routing")
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
```