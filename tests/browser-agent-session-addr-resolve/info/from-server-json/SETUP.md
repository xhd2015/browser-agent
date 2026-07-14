# Scenario

**Bug**: `session info` without `--addr` hits default `:43761` instead of meta addr

```
RunDaemon(:0) -> server.json
POST /v1/sessions
HandleCLI(["session","info","--session-id",id,"--base-dir",dir])   # NO --addr
  -> RED: 404 session not found
  -> GREEN: exit 0; stdout contains session_id
```

## Preconditions

- AddrSource = from-server-json.
- OmitAddr true (Run does not pass --addr).
- PassBaseDir true (from root Setup).

## Steps

1. Set AddrSource AddrFromServerJSON.
2. Ensure StartDaemon true; PassBaseDir true.

## Context

- Matches LOOP operator path: `session info --session-id X` with only `--base-dir`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.AddrSource = AddrFromServerJSON
	req.StartDaemon = true
	req.PassBaseDir = true
	return nil
}
```