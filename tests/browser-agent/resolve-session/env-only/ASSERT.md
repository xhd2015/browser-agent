## Expected

Requirement **A2** — only env set:

- Resolved id is `sess-from-env`.
- No error.

## Side Effects

- None.

## Errors

- Must not error when env is present.

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
	if resp.ResolvedID != "sess-from-env" {
		t.Fatalf("ResolvedID = %q, want %q", resp.ResolvedID, "sess-from-env")
	}
}
```
