## Expected

After implementer lands session list (**RED** on current code):

- `HandleCLI` returns nil; `ExitCode` 0.
- Stdout contains both `sess-alpha` and `sess-zulu`.
- `sess-alpha` appears before `sess-zulu` in stdout (sorted order).
- Stdout contains `2 sessions` (or equivalent count summary).
- Table headers mention `Session ID` and `Created` (and preferably `Status`).

## Side Effects

- Read-only list.

## Errors

- Missing session id, wrong sort order, or non-zero exit fails.

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
	if len(resp.CreatedSessionIDs) != 2 {
		t.Fatalf("harness created %d sessions want 2: %v", len(resp.CreatedSessionIDs), resp.CreatedSessionIDs)
	}

	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)

	want := []string{"sess-alpha", "sess-zulu"}
	for _, id := range want {
		if !strings.Contains(resp.Stdout, id) {
			t.Fatalf("stdout missing session %q; stdout=%s", id, truncate(resp.Stdout, 600))
		}
	}

	pos := stdoutSessionIDPositions(resp.Stdout, want)
	if pos["sess-alpha"] < 0 || pos["sess-zulu"] < 0 {
		t.Fatal("could not locate session ids in stdout")
	}
	if pos["sess-alpha"] >= pos["sess-zulu"] {
		t.Fatalf("stdout sort order wrong: alpha@%d zulu@%d; stdout=%s",
			pos["sess-alpha"], pos["sess-zulu"], truncate(resp.Stdout, 600))
	}

	assertContainsFold(t, resp.Stdout, "2 sessions", "session id", "created")
}```
