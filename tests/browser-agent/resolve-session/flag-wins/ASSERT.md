## Expected

Requirement **A1** — flag and env both set:

- `ResolveSessionID` returns `sess-from-flag`.
- No error.

## Side Effects

- None (pure).

## Errors

- Must not return env value when flag is set.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ResolveErr != "" {
		t.Fatalf("unexpected resolve error: %s", resp.ResolveErr)
	}
	if resp.ResolvedID != "sess-from-flag" {
		t.Fatalf("ResolvedID = %q, want %q (flag must win over env)",
			resp.ResolvedID, "sess-from-flag")
	}
}
```
