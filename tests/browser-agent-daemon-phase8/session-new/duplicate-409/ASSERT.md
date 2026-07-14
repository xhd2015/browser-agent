## Expected

- First `SessionNew` succeeds (`SessionNewErr` empty).
- Second `SessionNew` returns non-nil error (`SecondSessionNewErr` non-empty).
- Second error mentions **duplicate**, **exists**, or **409** (case-insensitive).
- Exactly one session `sess-dup-8` on server (at most one in list).

## Side Effects

- First create writes session dir; second does not add another registry entry.

## Errors

- Second call succeeding (no error) fails this leaf.

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
	if resp.SessionNewErr != "" {
		t.Fatalf("first SessionNew should succeed; got error: %s", resp.SessionNewErr)
	}
	if resp.SecondSessionNewErr == "" {
		t.Fatal("second SessionNew should fail with duplicate error")
	}
	low := strings.ToLower(resp.SecondSessionNewErr)
	if !strings.Contains(low, "duplicate") && !strings.Contains(low, "exists") && !strings.Contains(low, "409") {
		t.Fatalf("second error should mention duplicate/exists/409; got: %s", resp.SecondSessionNewErr)
	}
	count := 0
	for _, id := range resp.ServerSessionIDs {
		if id == req.SessionID {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("want exactly one %q on server; got count=%d ids=%v", req.SessionID, count, resp.ServerSessionIDs)
	}
}```
