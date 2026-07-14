# Scenario

**Feature**: CLI flag migration --host/--port/--server-port

```
serve --help / session new --help -> new flags
ResolveControlBaseURL -> server.json
```

## Preconditions

- Mode `ModeFlags`.
- Leaf sets op-specific field.

## Steps

1. Set `Mode = ModeFlags`.

## Context

- See root DOCTEST for `Run` dispatch.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeFlags
	return nil
}
```
