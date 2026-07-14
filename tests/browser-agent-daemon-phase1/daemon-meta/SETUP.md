# Scenario

**Feature**: daemon discovery file helpers for server.json

```
WriteDaemonMeta / ReadDaemonMeta / RemoveDaemonMeta on {BaseDir}/server.json
```

## Preconditions

- Mode is daemon-meta.
- Disk tests use temp BaseDir.

## Steps

1. Set Mode to daemon-meta.
2. Ensure BaseDir via ensureBaseDir.

## Context

- JSON fields: pid, addr, base_url, base_dir, started_at (RFC3339).
- Write is atomic (temp + rename) with trailing newline.

```go
import (
	"testing"

	"github.com/xhd2015/browser-agent/browseragent"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeDaemonMeta
	ensureBaseDir(t, req)
	return nil
}

func expectedSampleMeta(baseDir string) browseragent.DaemonMeta {
	return sampleDaemonMeta(baseDir)
}
```