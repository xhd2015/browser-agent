## Expected

Requirement **E1**:

- background.js found under `Chrome-Ext-Browser-Agent` (public/ or root/src/build).
- Source contains:
  - `/v1/ws` **or** `ws://`
  - `hello` (case-insensitive)
  - `job` (case-insensitive)
  - `result` (case-insensitive)
- Preferred (soft / best-effort): reconnect / onclose / retry / alarm tokens somewhere
  in file — **not required for GREEN** if core protocol present.

## Side Effects

- None (read-only FS).

## Errors

- Stub-only background without WS loop fails real extension agent.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.FileExists || strings.TrimSpace(resp.CombinedText) == "" {
		t.Fatalf("shell background missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	assertWSAgentTokens(t, resp.CombinedText, "shell background")
}
```
