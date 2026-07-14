## Expected

Requirement **E3**:

- Embedded background source found (package embed path or after extract).
- Same WS agent tokens as shell:
  - `/v1/ws` or `ws://`
  - hello, job, result (case-insensitive)

## Side Effects

- May extract under BaseDir when embed path missing (fallback).

## Errors

- Mini embed stub without protocol breaks load-unpacked from extract path.

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
		t.Fatalf("embedded background missing; err=%q found=%v ModuleRoot=%s",
			resp.ErrText, resp.FoundPaths, req.ModuleRoot)
	}
	assertWSAgentTokens(t, resp.CombinedText, "embedded background")
}
```
