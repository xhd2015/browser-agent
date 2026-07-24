## Expected

- `SessionNew` returns **nil** (timeout is soft: warning only).
- **Stdout** contains operator instructions:
  - session id `sess-wait-8`
  - `Session URL`
  - `browser-agent session info`
- **Stderr** contains wait progress then soft timeout:
  - `Waiting for extension`
  - `warning:` and `did not connect` (or `within`)

## Side Effects

- Session still registered on daemon after timeout.

## Errors

- Missing stdout markers or missing wait/timeout stderr fails.
- Non-nil SessionNew error fails (wait timeout must not fail the command).

## Exit Code

- Not asserted.

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
	if resp.SessionNewErr != "" {
		t.Fatalf("SessionNew error (timeout should be soft): %s", resp.SessionNewErr)
	}
	if resp.Stdout == "" {
		t.Fatal("stdout is empty; expected operator instructions before wait")
	}
	assertContainsFold(t, resp.Stdout, "sess-wait-8")
	assertContainsFold(t, resp.Stdout, "Session URL")
	assertContainsFold(t, resp.Stdout, "browser-agent session info")

	if resp.Stderr == "" {
		t.Fatal("stderr is empty; expected Waiting… then warning")
	}
	assertContainsFold(t, resp.Stderr, "Waiting for extension")
	low := strings.ToLower(resp.Stderr)
	if !strings.Contains(low, "warning:") {
		t.Fatalf("stderr missing warning: prefix; got:\n%s", truncate(resp.Stderr, 800))
	}
	if !strings.Contains(low, "did not connect") && !strings.Contains(low, "within") {
		t.Fatalf("stderr missing timeout wording; got:\n%s", truncate(resp.Stderr, 800))
	}
	// Instructions must not live only on stderr.
	if strings.Contains(strings.ToLower(resp.Stderr), "session url:") {
		t.Fatalf("Session URL should be on stdout before wait, not stderr; stderr=\n%s", truncate(resp.Stderr, 400))
	}
}
```
