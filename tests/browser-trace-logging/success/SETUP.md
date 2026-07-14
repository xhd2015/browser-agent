# Scenario

**Feature**: successful session path — logging on exit 0 with mock complete

```
# Happy path: hello + recording + complete
Mock Extension -> POST /v1/hello
Mock Extension -> status recording
Mock Extension -> POST /v1/complete
browser-trace -> exit 0
browser-trace stdout -> "{sessionDir}\n"
```

## Preconditions

- Extension script is record-and-complete; stop initiated by extension complete.
- `Quiet` / `Verbose` / `NoLogFile` narrowed by descendants.

## Steps

1. Set `ExtensionScript = record-and-complete`.
2. Set `StopMode = extension`.
3. Use moderate timeouts (product-scale is fine; mock completes quickly).

## Context

- All success leaves share exit 0 and exact stdout path contract.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtensionScript = ExtRecordAndComplete
	req.StopMode = StopExtension
	req.MockStopReason = "extension"
	// Fast enough for CI; mock does not need long ready window.
	if req.ReadyTimeout > 10*time.Second || req.ReadyTimeout == 30*time.Second {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.CompleteTimeout > 10*time.Second || req.CompleteTimeout == 30*time.Second {
		req.CompleteTimeout = 5 * time.Second
	}
	return nil
}
```
