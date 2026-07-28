# Scenario

**Feature**: SessionNew package API stamps durable session identity

```
SessionNew(Browser=…, Home=TestHome, NoOpenChrome, NoWait)
  -> meta.json + optional GET /v1/session reflect browser install tree
```

## Preconditions

- Mode = session-new.
- Ephemeral daemon + isolated TestHome / BaseDir.

## Steps

1. Set `Mode = ModeSessionNew`.
2. Leave Browser / op fields for child Setup.

## Context

- Open hooks must not run (NoOpenChrome in Run).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeSessionNew
	return nil
}
```
