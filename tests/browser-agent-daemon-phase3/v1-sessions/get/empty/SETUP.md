# Scenario

**Feature**: GET `/v1/sessions` on fresh server → empty array

```
Fresh registry server -> GET /v1/sessions -> []
```

## Preconditions

- No pre-created sessions.

## Steps

1. Leave `PreCreateSessionIDs` empty.

## Context

- Accept `[]` or empty JSON array.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	return nil
}
```