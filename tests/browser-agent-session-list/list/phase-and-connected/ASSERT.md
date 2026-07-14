## Expected

After implementer lands session-rich list columns (**RED** on current code):

- `HandleCLI` returns nil; `ExitCode` 0.
- Stdout contains `sess-wait-list` and `sess-conn-list`.
- Table headers include `Created`, `Pages`, `Browser`, and `Status`.
- Waiting session row shows unknown page count (`—`) or `unknown` status.
- Connected session row shows `ready` status when hello includes single-page telemetry.

## Side Effects

- Read-only list; fake extension stays connected during list.

## Errors

- Old Phase/Connected-only columns without new columns fails.

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
	if resp.WaitingSessionID == "" || resp.ConnectedSessionID == "" {
		t.Fatal("harness did not seed waiting and connected sessions")
	}

	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)

	assertContainsFold(t, resp.Stdout, "created", "pages", "browser", "status")
	assertContainsFold(t, resp.Stdout, resp.WaitingSessionID, resp.ConnectedSessionID)

	waitLine := lineContaining(resp.Stdout, resp.WaitingSessionID)
	connLine := lineContaining(resp.Stdout, resp.ConnectedSessionID)
	if waitLine == "" || connLine == "" {
		t.Fatal("session rows missing from stdout")
	}

	waitHasUnknown := strings.Contains(waitLine, "—") ||
		strings.Contains(waitLine, "--") ||
		(strings.Contains(strings.ToLower(waitLine), "unknown") &&
			!strings.Contains(strings.ToLower(resp.WaitingSessionID), "unknown"))
	if !waitHasUnknown {
		t.Fatalf("waiting session should show unknown pages/status; row=%q", waitLine)
	}

	assertContainsFold(t, connLine, "ready")
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