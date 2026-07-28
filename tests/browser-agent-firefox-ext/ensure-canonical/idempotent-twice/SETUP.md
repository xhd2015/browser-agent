# Scenario

**Feature**: second EnsureCanonicalFirefoxExtensionWithHome is idempotent

```
EnsureCanonicalFirefoxExtensionWithHome x2 -> same path + same version
```

## Preconditions

- EnsureCanonicalOp = idempotent-twice.

## Steps

1. Set EnsureCanonicalOp = EnsureCanonicalOpIdempotentTwice.

## Context

- Same package version must not create a second directory.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.EnsureCanonicalOp = EnsureCanonicalOpIdempotentTwice
	return nil
}
```
