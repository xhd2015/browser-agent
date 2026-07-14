# Scenario

**Feature**: shared React protocol jobs module exports job type constants

```
Test Client -> read react/src/lib/protocol/jobs.ts (or .js)
  must exist
  must contain all six type strings
```

## Preconditions

- Mode is protocol-src.
- ModuleRoot resolved by root Setup.

## Steps

1. Set `Mode = ModeProtocolSrc`.

## Context

- Requirement E1. Preferred path under react/src/lib/protocol/.

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
