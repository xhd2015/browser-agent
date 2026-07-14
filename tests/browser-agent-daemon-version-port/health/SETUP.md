# Scenario

**Feature**: extended GET /v1/health + server.json daemon_version

```
RunDaemon -> GET /v1/health {product, daemon_version, base_dir}
RunDaemon -> server.json daemon_version
```

## Preconditions

- Mode `ModeHealth`.
- Leaf sets op-specific field.

## Steps

1. Set `Mode = ModeHealth`.

## Context

- See root DOCTEST for `Run` dispatch.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeHealth
	return nil
}
```
