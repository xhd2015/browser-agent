# Scenario

**Feature**: CLI --tab-id / --tab-index flag parsing and job payload

```
HandleCLI session eval --tab-id N -> POST /v1/jobs includes tab_id
HandleCLI with both --tab-id and --tab-index -> exit 1 before POST
```

## Preconditions

- Mode is `cli`.
- `BaseDir` allocated per leaf for server-backed `tab-id-flag-posts-payload`.
- Conflict leaf needs no running daemon.

## Steps

1. Set `Mode = ModeCLI`.
2. Allocate temp `BaseDir` for payload leaf.
3. Leaf sets `CLIOp`.

## Context

- Mutual exclusion must surface before any job POST (no fake extension needed).

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeCLI
	dir := t.TempDir()
	req.BaseDir = filepath.Join(dir, "browser-agent-session-tab-targeting-cli")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	return nil
}
```