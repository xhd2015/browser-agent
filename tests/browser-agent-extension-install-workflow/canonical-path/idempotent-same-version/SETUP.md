# Scenario

**Feature**: second EnsureCanonicalExtension returns same path

```
EnsureCanonicalExtension() x2 -> same path + version
```

## Preconditions

- Same embedded version across calls.

## Steps

1. Set `CanonicalPathOp = idempotent-same-version`.

## Context

- No duplicate version directories.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CanonicalPathOp = CanonicalPathOpIdempotentSameVersion
	return nil
}
```