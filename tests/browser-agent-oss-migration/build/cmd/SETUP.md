# Scenario

**Feature**: `cmd/browser-agent` main builds

```
# compile CLI binary to /dev/null
Test Client -> go build -o /dev/null ./cmd/browser-agent/
Test Client <- success
```

## Preconditions

- `cmd/browser-agent/` entrypoint present after migration.

## Steps

1. Set `Leaf = cmd`.

## Context

- Discards binary via `/dev/null` (portable on darwin/linux harness).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Leaf = LeafBuildCmd
	return nil
}
```