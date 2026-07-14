# Scenario

**Feature**: explicit `--addr` overrides `server.json` (regression guard)

```
RunDaemon(:0) -> server.json
POST /v1/sessions
HandleCLI session info --session-id --base-dir --addr <meta base URL>
  -> exit 0 (may already GREEN before fix)
```

## Preconditions

- AddrSource = explicit-addr.
- Run passes `--addr` matching live daemon `BaseURL`.

## Steps

1. Set AddrSource AddrExplicit.
2. StartDaemon true; PassBaseDir true.

## Context

- Confirms resolution order: explicit `--addr` wins over meta discovery.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.AddrSource = AddrExplicit
	req.StartDaemon = true
	req.PassBaseDir = true
	return nil
}
```