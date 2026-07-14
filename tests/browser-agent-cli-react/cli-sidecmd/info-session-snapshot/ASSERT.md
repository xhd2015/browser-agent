## Expected

Requirement **B2**:

- HandleCLI nil error; ExitCode 0.
- Stdout ends with `\n`.
- Stdout contains the live session id (`resp.RealSessionID` or `req.SessionID`).
- Stdout mentions `connected` (JSON field or human label).

## Side Effects

- None required.

## Errors

- Empty stdout or missing session id fails.

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
		t.Fatal("info sidecmd timed out")
	}
	if resp.CLIErr != "" {
		t.Fatalf("info should succeed; CLIErr=%q stderr=%q stdout=%q",
			resp.CLIErr, resp.Stderr, resp.Stdout)
	}
	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)

	wantID := resp.RealSessionID
	if wantID == "" {
		wantID = req.SessionID
	}
	if !strings.Contains(resp.Stdout, wantID) {
		t.Fatalf("info stdout missing session id %q; stdout=%q", wantID, resp.Stdout)
	}
	if !strings.Contains(strings.ToLower(resp.Stdout), "connected") {
		t.Fatalf("info stdout should include connected field; stdout=%q", resp.Stdout)
	}
}
```
