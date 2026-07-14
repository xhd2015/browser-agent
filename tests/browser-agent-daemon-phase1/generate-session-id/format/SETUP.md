# Scenario

**Feature**: GenerateSessionID output matches format and varies across calls

```
GenerateSessionID() x2 -> ^sess-[a-z0-9]{6}$ and ids differ
```

## Preconditions

- Mode generate-session-id (parent).

## Steps

1. Run calls GenerateSessionID twice.

## Context

- Probabilistic uniqueness: two calls should differ (retry in assert if needed).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	return nil
}
```