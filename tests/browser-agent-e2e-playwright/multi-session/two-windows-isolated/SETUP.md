# Scenario

**Feature**: Two windows/tabs — isolated session connections

```
harness POST sess-e2e-a + sess-e2e-b
playwright-debug -> tab A /go?session=a, tab B /go?session=b
wait both extension.connected
```

## Preconditions

- `PlaywrightOp` two-windows-isolated.
- Session ids `sess-e2e-a`, `sess-e2e-b`.

## Steps

1. Set `PlaywrightOp = PlaywrightOpTwoWindowsIsolated`.
2. Set `SessionID = "sess-e2e-a"`, `SessionIDB = "sess-e2e-b"`.

## Context

- Script: `testdata/two-windows-isolated.js`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.PlaywrightOp = PlaywrightOpTwoWindowsIsolated
	req.SessionID = "sess-e2e-a"
	req.SessionIDB = "sess-e2e-b"
	return nil
}
```