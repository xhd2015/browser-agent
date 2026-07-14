# Scenario

**Feature**: SessionNew extracts canonical extension before opening Chrome

```
SessionNew -> EnsureCanonicalExtension dir exists on disk
```

## Preconditions

- Fresh `TestHome` before SessionNew.

## Steps

1. Set `SessionNewOp = extracts-before-open`.

## Context

- `ExtensionPath` populated from post-SessionNew `EnsureCanonicalExtension` probe.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionNewOp = SessionNewOpExtractsBeforeOpen
	return nil
}
```