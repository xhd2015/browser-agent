## Expected

- `SessionNew` returns **nil** error.
- Stdout contains created session id `sess-pretty-8`.
- Stdout contains export hint: `BROWSER_AGENT_SESSION_ID` and/or `export`.
- Stdout contains inspect/interact markers:
  - `browser-agent session info`
  - at least one of `eval`, `run`, `logs`, `screenshot`, `cdp` in nested form.

## Side Effects

- Formatter writes only to provided `Stdout` writer.

## Errors

- Missing markers or session new error fails.

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
		t.Fatalf("SessionNew error: %s", resp.SessionNewErr)
	}
	if resp.Stdout == "" {
		t.Fatal("stdout is empty")
	}
	assertContainsFold(t, resp.Stdout, req.SessionID)
	assertContainsFold(t, resp.Stdout, "browser-agent session info")
	// Export hint for shell sessions.
	if !containsAnyFold(resp.Stdout, "browser_agent_session_id", "export") {
		t.Fatalf("stdout missing export hint (BROWSER_AGENT_SESSION_ID or export); got:\n%s", truncate(resp.Stdout, 800))
	}
	// At least one interact recipe beyond info.
	if !containsAnyFold(resp.Stdout, "session eval", "session run", "session logs", "session screenshot", "session cdp") {
		t.Fatalf("stdout missing interact/eval hints; got:\n%s", truncate(resp.Stdout, 800))
	}
}

func containsAnyFold(haystack string, needles ...string) bool {
	low := strings.ToLower(haystack)
	for _, n := range needles {
		if strings.Contains(low, strings.ToLower(n)) {
			return true
		}
	}
	return false
}```
