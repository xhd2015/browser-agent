# Scenario

**Feature**: session directory path and existence helpers

```
SessionDirPath(baseDir, id) + SessionDirExists(baseDir, id)
```

## Preconditions

- Mode is session-dir.
- Disk tests use temp BaseDir.

## Steps

1. Set Mode to session-dir.
2. Ensure BaseDir via ensureBaseDir.

## Context

- Path shape: `{baseDir}/sessions/{sessionID}`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSessionDir
	ensureBaseDir(t, req)
	if req.SessionDirID == "" {
		req.SessionDirID = "my-flow"
	}
	return nil
}

func expectedSessionDirPath(baseDir, sessionID string) string {
	return filepath.Join(baseDir, "sessions", sessionID)
}
```