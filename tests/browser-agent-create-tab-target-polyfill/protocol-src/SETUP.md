# Scenario

**Feature**: shared React protocol jobs module exports create_tab

```
Test Client -> read react/src/lib/protocol/jobs.ts (or .js)
  must exist
  must contain create_tab / JOB_TYPE_CREATE_TAB
```

## Preconditions

- Mode is protocol-src.
- ModuleRoot resolved by root Setup.

## Steps

1. Set `Mode = ModeProtocolSrc`.

## Context

- Requirement P1. Additive token; sealed six-token asserts remain valid.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeProtocolSrc
	return nil
}
```
