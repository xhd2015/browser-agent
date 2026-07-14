# Scenario

**Feature**: list resolves daemon addr from server.json when --addr omitted

```
RunDaemon(:0, BaseDir) -> server.json (addr != 43761)
POST /v1/sessions -> sess-json-list
HandleCLI session list --base-dir BaseDir   # no --addr
  -> exit 0; session id in stdout
```

## Preconditions

- Ephemeral listen port recorded in meta; `--addr` intentionally omitted.

## Steps

1. Set `ListOp = ListOpFromServerJSON`.
2. `SessionIDsToCreate = ["sess-json-list"]`.
3. `OmitAddr = true`.

## Context

- Reproduces session-addr-resolve pattern for list sub-command.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ListOp = ListOpFromServerJSON
	req.SessionIDsToCreate = []string{"sess-json-list"}
	req.OmitAddr = true
	req.JSONMode = false
	return nil
}```
