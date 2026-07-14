# Scenario

**Feature**: delete rejected when extension connected

```
POST /v1/sessions -> Fake Extension hello -> extension_connected
HandleCLI session delete -> exit 1; extension connected; session remains
```

## Preconditions

- Fake extension WS hello before delete attempt.

## Steps

1. Set `CLIOp = connected-rejected`.
2. `ConnectExtension = true`.

## Context

- Extension socket stays open during delete.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIOp = CLIOpConnectedReject
	req.ConnectExtension = true
	return nil
}
```