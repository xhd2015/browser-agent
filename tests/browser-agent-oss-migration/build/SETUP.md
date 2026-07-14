# Scenario

**Feature**: migrated repo compiles browser-agent packages

```
# go build from repo root
Test Client -> go build <package pattern>
Test Client <- exit 0
```

## Preconditions

- Module path and imports updated to `github.com/xhd2015/browser-agent`.
- `go` on PATH.

## Steps

1. Set `Category = build`.

## Context

- Builds run with `cmd.Dir = RepoRoot`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Category = CategoryBuild
	return nil
}
```