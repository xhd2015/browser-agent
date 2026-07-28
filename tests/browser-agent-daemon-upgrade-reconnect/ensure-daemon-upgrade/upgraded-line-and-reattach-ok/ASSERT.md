## Expected

- EnsureDaemon ok; kill+spawn called.
- Stderr contains `upgraded daemon` plus old `0.1.0` and new `0.2.0` (arrow `→` or `->` optional).
- Stderr mentions reattach and both session ids `sess-r1`, `sess-r2`.

## Side Effects

- None.

## Errors

- Missing upgraded line or reattach list fails.

## Exit Code

- 0.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.EnsureErr != nil {
		t.Fatalf("EnsureDaemon err=%v want nil", resp.EnsureErr)
	}
	if !resp.KillFnCalled || !resp.SpawnFnCalled {
		t.Fatal("expected kill+spawn on upgrade")
	}
	assertContainsFold(t, resp.Stderr, "upgraded daemon", "0.1.0", "0.2.0")
	// Prefer unicode arrow but accept ASCII.
	if !strings.Contains(resp.Stderr, "→") && !strings.Contains(resp.Stderr, "->") {
		// Soft: some implementations may use "to" / "->"; upgraded + versions is required above.
		// If neither arrow form present, still require versions adjacent-ish via upgraded line.
	}
	assertContainsFold(t, resp.Stderr, "reattach")
	assertContainsFold(t, resp.Stderr, "sess-r1", "sess-r2")
	assertExitZero(t, resp)
}
```
