# Scenario

**Feature**: when `server.json` is absent, side-commands fall back to default `:43761`

```
# no RunDaemon in BaseDir
HandleCLI session info --session-id --base-dir   # no --addr, no meta
  -> uses http://127.0.0.1:43761
  -> connection refused or health fail OK
```

## Preconditions

- No daemon started in leaf `BaseDir`.
- `server.json` must not exist under `BaseDir`.

## Steps

1. Child leaf sets AddrSource default-no-meta.
2. StartDaemon false for fallback leaves.

## Context

- Documents resolution order step 3; not the primary bug repro.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.StartDaemon = false
	return nil
}
```