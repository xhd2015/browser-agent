# Scenario

**Feature**: `go test ./browseragent/...` passes

```
# unit test all browseragent packages
Test Client -> go test ./browseragent/...
Test Client <- exit 0
```

## Preconditions

- Test files import `github.com/xhd2015/browser-agent/browseragent`.

## Steps

1. Set `Leaf = browseragent-unit`.

## Context

- Fails RED until migration + module rename complete.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Leaf = LeafBrowseragentUnit
	return nil
}
```