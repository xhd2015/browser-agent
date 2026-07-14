## Expected

After implementer lands session-rich (**RED** on current code):

- `HandleCLI` returns nil; `ExitCode` 0.
- Stdout contains session id.
- Pages column shows em dash `—` (or `unknown` status label) for unknown telemetry.
- Stdout does **not** show page count `0` for this session (unknown ≠ zero).

## Side Effects

- Read-only list; no extension connection.

## Errors

- Showing `0` instead of `—` for unknown count fails.

## Exit Code

- 0.

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
	if resp.DispatchTimedOut {
		t.Fatal("session list timed out")
	}

	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)
	assertContainsFold(t, resp.Stdout, req.SessionID)

	line := lineContaining(resp.Stdout, req.SessionID)
	if line == "" {
		t.Fatalf("session row missing for %q; stdout=%s", req.SessionID, truncate(resp.Stdout, 600))
	}

	hasDash := strings.Contains(line, "—") || strings.Contains(line, "--")
	// Status column may show "unknown" label; do not match session id substrings.
	hasUnknownStatus := strings.Contains(strings.ToLower(line), "unknown") &&
		!strings.Contains(strings.ToLower(req.SessionID), "unknown")
	if !hasDash && !hasUnknownStatus {
		t.Fatalf("Pages column should show — or unknown status; row=%q stdout=%s", line, truncate(resp.Stdout, 600))
	}

	// Must not mislabel unknown as zero pages in the row.
	if strings.Contains(line, " 0 ") || strings.HasSuffix(strings.TrimSpace(line), " 0") {
		t.Fatalf("unknown page count must not display as 0; row=%q", line)
	}
}

func lineContaining(text, needle string) string {
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(strings.ToLower(line), strings.ToLower(needle)) {
			return line
		}
	}
	return ""
}
```