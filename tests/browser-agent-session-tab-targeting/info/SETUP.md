# Scenario

**Feature**: session info human table + --json tab index fields

```
RunDaemon -> fake extension (info job tabs + job_target)
HandleCLI session info [--json] -> enriched tab targeting output
```

## Preconditions

- Mode is `info`.
- Temp `BaseDir` per leaf.
- Fake extension auto-completes info job with tab index fixtures.

## Steps

1. Set `Mode = ModeInfo`.
2. Allocate temp `BaseDir`.
3. Leaf sets `InfoOp`.

## Context

- Reuses RunDaemon + fake extension pattern from `browser-agent-session-rich`.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeInfo
	dir := t.TempDir()
	req.BaseDir = filepath.Join(dir, "browser-agent-session-tab-targeting-info")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	return nil
}
```