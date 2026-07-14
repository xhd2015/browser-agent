# Scenario

**Feature**: no `server.json` + no `--addr` → default control addr (connection error OK)

```
# BaseDir has no server.json
HandleCLI session info --session-id sess-fallback --base-dir BaseDir
  -> normalizeAddr("") -> http://127.0.0.1:43761
  -> connection error or remote 404 for unknown id (not meta-resolved ephemeral port)
```

## Preconditions

- AddrSource = default-no-meta.
- StartDaemon false.
- PassBaseDir true.

## Steps

1. Set Sidecmd SidecmdInfo.
2. Set AddrSource AddrDefaultNoMeta.
3. Set StartDaemon false.

## Context

- Regression for resolution order; does not require a live daemon on 43761.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Sidecmd = SidecmdInfo
	req.AddrSource = AddrDefaultNoMeta
	req.StartDaemon = false
	req.PassBaseDir = true
	return nil
}
```