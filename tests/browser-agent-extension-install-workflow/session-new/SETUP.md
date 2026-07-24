# Scenario

**Feature**: SessionNew system Chrome + canonical extract workflow

```
RunDaemon -> SessionNew(OpenChromeFn=record, NoWait) -> system argv + enriched stdout
```

## Preconditions

- Ephemeral daemon on `:0`; temp `BaseDir`.
- `TestHome` for canonical path assertions.
- `OpenChromeFn` records system-chrome argv (`BuildChromeArgs(url, "")`).
- `NoWait` avoids 30s extension wait without a real browser.
## Steps

1. Set `Mode = session-new`.
2. Leaf sets `SessionNewOp` and optional `NoOpenChrome`.

## Context

- Never invokes agent-run. Parallel-safe via per-config `OpenChromeFn`.

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