## Expected

- Fake extension receives a job with type `eval`.
- WS job envelope includes `tab_id` equal to **216771574**.
- `HandleCLI` succeeds (`ExitCode` 0, empty `CLIErr`).

## Side Effects

- One eval job enqueued and pushed over WS.

## Errors

- Missing `tab_id`, wrong value, or ignored `--tab-id` flag fails.

## Exit Code

- **0**.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.DispatchTimedOut {
		t.Fatal("session eval --tab-id timed out")
	}
	if resp.CLIErr != "" {
		t.Fatalf("session eval --tab-id should succeed; CLIErr=%q stderr=%q stdout=%q",
			resp.CLIErr, truncate(resp.Stderr, 300), truncate(resp.Stdout, 300))
	}
	if resp.ExitCode != 0 {
		t.Fatalf("ExitCode=%d want 0; cliErr=%q", resp.ExitCode, resp.CLIErr)
	}
	if !resp.WSJobReceived {
		t.Fatalf("fake extension did not observe a job; raw=%s", truncate(resp.ObservedJobRaw, 400))
	}
	if resp.ObservedJobType != "eval" {
		t.Fatalf("ObservedJobType=%q want eval; raw=%s", resp.ObservedJobType, truncate(resp.ObservedJobRaw, 400))
	}
	wantTabID := int64(216771574)
	if resp.ObservedTabID != wantTabID {
		t.Fatalf("ObservedTabID=%d want %d; raw=%s", resp.ObservedTabID, wantTabID, truncate(resp.ObservedJobRaw, 600))
	}
}
```