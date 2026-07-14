## Expected

- `SessionNew` succeeds.
- Stdout contains `session info`, `session eval`, `session run` recipe lines.
- Stdout contains `session-id:` and `export BROWSER_AGENT_SESSION_ID`.

## Side Effects

- None.

## Errors

- Missing legacy recipe markers fails.

## Exit Code

- Not asserted.

```go
import (
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
	assertContainsFold(t, resp.Stdout,
		"session-id:",
		"export browser_agent_session_id",
		"session info",
		"session eval",
		"session run",
	)
}
```