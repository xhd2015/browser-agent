# Scenario

**Feature**: Real-browser E2E tab_id targeting via job POST (CLI contract)

```
RunDaemon -> POST /v1/sessions -> playwright-debug --extension --headed
Multi-tab window -> POST jobs with tab_id -> extension targets pinned tab
```

## Preconditions

- Mode is `e2e`.
- `playwright-debug` on PATH; skip when absent.
- Default `ReadyTimeout = 10s`, `PlaywrightTimeout = 90s`.

## Steps

1. Set `Mode = ModeE2E`.
2. Allocate temp `BaseDir`.
3. Leaf sets `PlaywrightOp` and `SessionID`.

## Context

- Playwright scripts POST `/v1/jobs` with `tab_id` (same payload CLI `--tab-id` produces).
- ASSERT frontmatter: `slow, ui-automation`.

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
	req.BaseDir = filepath.Join(dir, "browser-agent-session-tab-targeting-e2e")
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