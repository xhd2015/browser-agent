# Scenario

**Feature**: browseragent unit tests pass after migration

```
# go test from repo root
Test Client -> go test ./browseragent/...
Test Client <- all packages pass
```

## Preconditions

- Unit tests under `browseragent/` copied and import paths updated.

## Steps

1. Set `Category = test`.

## Context

- No integration Chrome required for package unit tests.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Category = CategoryTest
	return nil
}
```