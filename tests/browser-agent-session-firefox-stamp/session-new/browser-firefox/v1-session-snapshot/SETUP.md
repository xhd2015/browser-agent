# Scenario

**Feature**: GET /v1/session exposes firefox extension_install_path after stamp

```
SessionNew(Browser=firefox) -> GET /v1/session?session=<id>
  -> extension_install_path contains browser-agent-firefox
```

## Preconditions

- SessionNewFirefoxOp = v1-session-snapshot.
- FetchV1Session enabled.

## Steps

1. Set SessionNewFirefoxOp = SessionNewFirefoxOpV1SessionSnapshot.
2. Set FetchV1Session = true.

## Context

- Live snapshot must match durable stamp for injectSessionBoot / SPA fallbacks.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionNewFirefoxOp = SessionNewFirefoxOpV1SessionSnapshot
	req.FetchV1Session = true
	return nil
}
```
