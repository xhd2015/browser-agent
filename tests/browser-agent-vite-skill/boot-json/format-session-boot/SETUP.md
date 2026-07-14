# Scenario

**Feature**: FormatSessionBootJSON fields for React boot config (D1)

```
FormatSessionBootJSON("boot-sess-fixed")
  -> session_id == boot-sess-fixed
  -> product == browser-agent
  -> control_port == 43761
```

## Preconditions

- BootSessionID fixed for deterministic assert.

## Steps

1. Set BootSessionID to a known value.

## Context

- Extra JSON fields allowed.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.BootSessionID = "boot-sess-fixed"
	return nil
}
```
