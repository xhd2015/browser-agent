# Scenario

**Feature**: `browseragent` package tree builds

```
# compile all browseragent packages
Test Client -> go build ./browseragent/...
Test Client <- success
```

## Preconditions

- OSS core package `browseragent/` present after migration.

## Steps

1. Set `Leaf = browseragent`.

## Context

- Uses Go recursive package pattern.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Leaf = LeafBuildBrowseragent
	return nil
}
```