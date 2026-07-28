# Scenario

**Feature**: RestoreSessionsFromDisk rehydrates registry from on-disk session dirs

```
seed {baseDir}/sessions/*/meta.json
NewSessionRegistry -> RestoreSessionsFromDisk
  -> List/Get registered ids (waiting_extension) or empty when skipped
```

## Preconditions

- Mode = restore-from-disk.
- Seeds prepared by leaf Setup under temp BaseDir.

## Steps

1. Set `Mode = ModeRestoreFromDisk`.
2. Leaves set `RestoreCase` and `Seeds`.

## Context

- Best-effort skip for invalid/corrupt entries; happy path registers waiting_extension.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeRestoreFromDisk
	return nil
}
```
