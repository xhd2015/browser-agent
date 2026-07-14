# Scenario

**Feature**: pure FormatSessionBootJSON helper (D)

```
Test Client -> FormatSessionBootJSON(sessionID)
  -> JSON {session_id, product:browser-agent, control_port:43761}
```

## Preconditions

- Mode = ModeBootJSON.
- No server / no HTTP / no Chrome.

## Steps

1. Set Mode = ModeBootJSON.
2. Leaf sets BootSessionID.

## Context

- Pure unit-of-behavior via package helper used by HTML inject path.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeBootJSON
	return nil
}
```
