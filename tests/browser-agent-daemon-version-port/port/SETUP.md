# Scenario

**Feature**: fixed default control port + host/port bind

```
serve --host/--port -> bind 127.0.0.1:N
EnsureDaemon spawn -> default port (no :0)
```

## Preconditions

- Mode `ModePort`.
- Leaf sets op-specific field.

## Steps

1. Set `Mode = ModePort`.

## Context

- See root DOCTEST for `Run` dispatch.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModePort
	return nil
}
```
