## Expected

Requirement **F1**:

- For each of `info`, `eval`, `run`, `logs`, `screenshot`, `cdp`:
  `IsKnownJobType` returns **true**.
- For unknowns (`""`, `foo`, `Eval`, `navigate`, `job`, `unknown-type`):
  returns **false**.
- Helper must be exported as `browseragent.IsKnownJobType` (compile-time).

## Side Effects

- None.

## Errors

- Accepting unknown types or rejecting known types fails shared validation.

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
	known := []string{"info", "eval", "run", "logs", "screenshot", "cdp"}
	for _, k := range known {
		ok, present := resp.KnownResults[k]
		if !present {
			t.Fatalf("KnownResults missing key %q", k)
		}
		if !ok {
			t.Fatalf("IsKnownJobType(%q) = false, want true", k)
		}
	}
	for u, ok := range resp.UnknownResults {
		if ok {
			t.Fatalf("IsKnownJobType(%q) = true, want false", u)
		}
	}
}
```
