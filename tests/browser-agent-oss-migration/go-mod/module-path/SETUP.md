# Scenario

**Feature**: `go.mod` module path is `github.com/xhd2015/browser-agent`

```
# first line of go.mod
Test Client -> open RepoRoot/go.mod
Test Client <- module github.com/xhd2015/browser-agent
```

## Preconditions

- Locked decision #1 in REQUIREMENT-DESIGN.

## Steps

1. Set `Leaf = module-path`.

## Context

- Assert exact first-line match (trimmed).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Leaf = LeafModulePath
	return nil
}
```