# Scenario

**Bug**: `session info` without `--addr` must resolve control URL from `server.json`

```
RunDaemon -> server.json (ephemeral addr)
HandleCLI session info --session-id --base-dir   # omit --addr
  -> GET correct /v1/session (after fix)
```

## Preconditions

- Sidecmd = info for this subtree.
- Daemon leaves start RunDaemon and create a session.

## Steps

1. Set Sidecmd to SidecmdInfo.
2. Child leaves set AddrSource and explicit-addr flags.

## Context

- Primary bug repro surface from LOOP session-info-addr-mismatch.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Sidecmd = SidecmdInfo
	return nil
}
```