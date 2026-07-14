# Scenario

**Feature**: Create rejects invalid or duplicate session ids

```
Create(id) -> validation error | ErrSessionExists
```

## Preconditions

- CreateCase set by leaf to an error variant.

## Steps

1. Inherit create grouping (Mode, BaseDir, Addr).

## Context

- Invalid id uses ValidateSessionID error (not ErrSessionExists).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	return nil
}
```