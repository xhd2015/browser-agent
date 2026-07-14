## Expected

After implementer lands session-rich (**RED** on current code):

- `HandleCLI` returns nil; `ExitCode` 0.
- Stdout is JSON object ending with `\n`.
- JSON includes non-empty `created_at`.
- JSON includes `status` (expected `ready` when connected with 1 page).
- JSON includes `session_page_count` equal to `1`.

## Side Effects

- Fake extension connected during info.

## Errors

- Missing enriched fields or human-only output fails.

## Exit Code

- 0.

```go
import (
	"strings"
	"testing"
	"time"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.DispatchTimedOut {
		t.Fatal("session info --json timed out")
	}

	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)

	if !stdoutLooksLikeJSONObject(resp.Stdout) {
		t.Fatalf("stdout should be JSON object with --json; got %s", truncate(resp.Stdout, 400))
	}

	if resp.CreatedAt == "" {
		t.Fatalf("created_at missing from --json stdout; stdout=%s", truncate(resp.Stdout, 600))
	}
	if _, err := time.Parse(time.RFC3339, resp.CreatedAt); err != nil {
		if _, err2 := time.Parse(time.RFC3339Nano, resp.CreatedAt); err2 != nil {
			t.Fatalf("created_at=%q not parseable: %v", resp.CreatedAt, err)
		}
	}

	if resp.Status == "" {
		t.Fatalf("status missing from --json stdout; stdout=%s", truncate(resp.Stdout, 600))
	}
	if resp.Status != "ready" {
		t.Fatalf("status=%q want ready for connected 1-page session; stdout=%s",
			resp.Status, truncate(resp.Stdout, 600))
	}

	if resp.SessionPageCount == nil || *resp.SessionPageCount != 1 {
		got := -1
		if resp.SessionPageCount != nil {
			got = *resp.SessionPageCount
		}
		t.Fatalf("session_page_count=%d want 1; stdout=%s", got, truncate(resp.Stdout, 600))
	}

	low := strings.ToLower(resp.Stdout)
	if !strings.Contains(low, "created_at") || !strings.Contains(low, "session_page_count") {
		t.Fatalf("stdout missing enriched keys; stdout=%s", truncate(resp.Stdout, 600))
	}
}
```