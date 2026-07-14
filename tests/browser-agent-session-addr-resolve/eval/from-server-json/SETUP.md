# Scenario

**Bug**: `session eval` without `--addr` gets 404 unknown session on wrong host

```
RunDaemon(:0) -> server.json
POST /v1/sessions
HandleCLI session eval --session-id --base-dir '1+1'   # NO --addr
  -> RED: status 404 / session not found
  -> GREEN: must NOT contain unknown session / session not found
```

## Preconditions

- AddrSource = from-server-json.
- No `--addr` on argv.
- Eval expression `1+1`.

## Steps

1. Set AddrSource AddrFromServerJSON.
2. StartDaemon true; PassBaseDir true.

## Context

- Job may timeout or fail waiting for extension — acceptable after fix.
- Failure mode to reject: HTTP 404 `unknown session id` from wrong control host.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.AddrSource = AddrFromServerJSON
	req.StartDaemon = true
	req.PassBaseDir = true
	req.EvalExpr = "1+1"
	return nil
}
```