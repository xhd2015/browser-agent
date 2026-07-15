# Scenario

**Feature**: jobs protocol module includes create_tab constant (P1)

```
react/src/lib/protocol/jobs.ts|js
  tokens: "create_tab" and/or JOB_TYPE_CREATE_TAB
```

## Preconditions

- Mode already protocol-src from parent.

## Steps

1. Ensure Mode stays ModeProtocolSrc.

## Context

- Requirement P1.

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
