## Expected

After implementer lands session list (**RED** on current code):

- `HandleCLI` returns nil; `ExitCode` 0.
- `server.json` exists; daemon addr is **not** default `:43761`.
- Stdout contains live `sess-json-list`.
- Must **not** fail with connection error to `127.0.0.1:43761` or `session not found`.

## Side Effects

- Read-only list via meta-resolved base URL.

## Errors

- 404/connection errors, missing session id, or `--addr` default-port bug fails.

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
	if !resp.DaemonMetaExists {
		t.Fatal("server.json must exist for from-server-json leaf")
	}
	if resp.DaemonMeta.Addr == "127.0.0.1:43761" || strings.HasSuffix(resp.DaemonMeta.Addr, ":43761") {
		t.Fatalf("ephemeral daemon must not bind :43761; meta=%+v", resp.DaemonMeta)
	}
	if len(resp.CreatedSessionIDs) != 1 {
		t.Fatalf("harness created %d sessions want 1", len(resp.CreatedSessionIDs))
	}
	sid := resp.CreatedSessionIDs[0]

	if resp.CLIErr != "" {
		t.Fatalf("session list without --addr should succeed via server.json; CLIErr=%q stderr=%q stdout=%q",
			resp.CLIErr, resp.Stderr, resp.Stdout)
	}
	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)

	if !strings.Contains(resp.Stdout, sid) {
		t.Fatalf("stdout missing session_id %q; stdout=%s", sid, truncate(resp.Stdout, 500))
	}

	out := combinedOutput(resp)
	assertNotContainsFold(t, out, "127.0.0.1:43761", "connection refused", "session not found")
}```
