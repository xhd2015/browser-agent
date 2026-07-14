# Scenario

**Feature**: GET /v1/session extension_install_path uses canonical layout

```
SessionNew -> GET /v1/session -> extension_install_path contains extensions/browser-agent/
```

## Preconditions

- Daemon running; session created via SessionNew.
- `NoOpenChrome` true in Run to avoid launch noise.

## Steps

1. Set `Mode = snapshot`.
2. Leaf sets `SnapshotOp`.

## Context

- Snapshot field set at session create time.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSnapshot
	return nil
}
```