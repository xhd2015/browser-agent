## Expected

Requirement **G1** (additive):

- `IsKnownJobType("create_tab")` returns **true**.
- Prior six still true: `info`, `eval`, `run`, `logs`, `screenshot`, `cdp`.
- Unknowns remain false (including hyphen CLI form `create-tab`, `CreateTab`, empty, foo).
- Helper exported as `browseragent.IsKnownJobType` (compile-time).
- **Does not** assert exclusive set size / len==7.

## Side Effects

- None.

## Errors

- Rejecting `create_tab` or prior six, or accepting hyphen/unknown forms, fails.

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
	if !resp.HelperAvailable {
		t.Fatal("IsKnownJobType helper not available")
	}
	// Additive known set — never exclusive count.
	known := []string{"info", "eval", "run", "logs", "screenshot", "cdp", "create_tab"}
	for _, k := range known {
		ok, present := resp.KnownResults[k]
		if !present {
			t.Fatalf("KnownResults missing key %q", k)
		}
		if !ok {
			t.Fatalf("IsKnownJobType(%q) = false, want true (additive create_tab / prior six)", k)
		}
	}
	for u, ok := range resp.UnknownResults {
		if ok {
			t.Fatalf("IsKnownJobType(%q) = true, want false", u)
		}
	}
}
```
