# Scenario

**Feature**: /v1/session while not connected — install path + guidance (requirement #5)

```
# No POST /v1/hello
Control Server Session: phase waiting_extension, connected=false
Test Client -> GET /v1/session?session=<id>
Control Server -> {
  extension_install_path: absolute,
  embedded_version: non-empty,
  extension.connected: false,
  hint: Load unpacked / path / chrome://extensions guidance
}
```

## Preconditions

- Live session after extract-on-start.
- No hello, no status.

## Steps

1. Set `DoHello = false`.

## Context

- User should not need to re-run a CLI install flag first — path is already in JSON.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DoHello = false
	return nil
}
```
