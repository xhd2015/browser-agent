# Scenario

**Feature**: `session new` waits for extension connection with different outcomes

```
Fake Extension -> hello { version, features }
Poll GET /v1/session?session=ID -> extension.connected, extension.features
  -> supported → stderr "Extension connected", exit 0
  -> unsupported → error, exit 1
  -> timeout → stderr warning, exit 0, stdout has session output
  -> daemon dies → error, exit 1
```

## Preconditions

- Mode is `with-wait`.
- Daemon is running, session created.
- Extension polls `GET /v1/session` every 500ms.

## Steps

1. Set `Mode = ModeWithWait`.
2. Leaves set `WaitOp` to select the extension outcome.

## Context

- Wait loop polls `GET /v1/session` until extension connects or timeout.
- Extension support check: version ≥ 1.0.0 AND features include `browser-agent`.
- Default timeout: 30s; test may use shorter timeout for faster feedback.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeWithWait
	return nil
}
```
