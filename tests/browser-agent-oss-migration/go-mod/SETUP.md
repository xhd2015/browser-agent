# Scenario

**Feature**: migrated repo declares OSS module path

```
# go.mod module line
Test Client -> read go.mod at repo root
Test Client <- first line module github.com/xhd2015/browser-agent
```

## Preconditions

- `go.mod` must exist at repo root after migration.

## Steps

1. Set `Category = go-mod`.

## Context

- Pure file read; no `go` subprocess.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Category = CategoryGoMod
	return nil
}
```