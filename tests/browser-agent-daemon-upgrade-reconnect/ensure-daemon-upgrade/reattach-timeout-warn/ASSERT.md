## Expected

- EnsureDaemon succeeds despite incomplete reattach (soft warn).
- Kill+spawn still occurred; upgraded line present.
- Stderr warns about still waiting / waiting reattach and includes waitList ids.

## Side Effects

- None.

## Errors

- Hard-failing EnsureDaemon on reattach timeout fails this leaf.
- Missing still-waiting warning fails.

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
		t.Fatalf("EnsureDaemon err=%v want nil (reattach timeout is soft warn)", resp.EnsureErr)
	}
	if !resp.KillFnCalled || !resp.SpawnFnCalled {
		t.Fatal("upgrade kill+spawn required even when reattach times out")
	}
	assertContainsFold(t, resp.Stderr, "upgraded daemon", req.DaemonVersion, req.ClientVersion)

	low := strings.ToLower(resp.Stderr)
	hasStill := strings.Contains(low, "still waiting") ||
		(strings.Contains(low, "waiting") && (strings.Contains(low, "reattach") || strings.Contains(low, "extension")))
	if !hasStill {
		t.Fatalf("stderr missing still-waiting warning; got:\n%s", truncate(resp.Stderr, 900))
	}
	assertContainsFold(t, resp.Stderr, "sess-w1", "sess-w2")
	assertExitZero(t, resp)
}
```
