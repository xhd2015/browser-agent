## Expected

After implementer lands session list (**RED** on current code):

- `HandleCLI` returns nil; `ExitCode` 0.
- Stdout parses as JSON **array** with one element.
- Element contains `session_id`, `phase`, and `extension.connected` fields.
- Stdout does **not** contain human table headers (`Session ID` column prose).
- Stdout contains no ANSI escape sequences.

## Side Effects

- Read-only JSON dump.

## Errors

- Non-JSON stdout, object wrapper instead of array, or table prose fails.

## Exit Code

- 0.

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.DispatchTimedOut {
		t.Fatal("session list --json timed out")
	}
	if len(resp.CreatedSessionIDs) != 1 {
		t.Fatalf("harness created %d sessions want 1", len(resp.CreatedSessionIDs))
	}
	sid := resp.CreatedSessionIDs[0]

	assertExitZero(t, resp)

	trim := strings.TrimSpace(resp.Stdout)
	if strings.Contains(trim, "Session ID") {
		t.Fatalf("--json must not emit human table; stdout=%s", truncate(resp.Stdout, 500))
	}
	if strings.Contains(resp.Stdout, "\x1b[") {
		t.Fatalf("--json must not emit ANSI; stdout=%s", truncate(resp.Stdout, 500))
	}

	var list []map[string]any
	if err := json.Unmarshal([]byte(trim), &list); err != nil {
		t.Fatalf("stdout should be JSON array; parse err=%v stdout=%s", err, truncate(resp.Stdout, 500))
	}
	if len(list) != 1 {
		t.Fatalf("JSON array len=%d want 1; stdout=%s", len(list), truncate(resp.Stdout, 500))
	}
	el := list[0]
	if got, _ := el["session_id"].(string); got != sid {
		t.Fatalf("JSON session_id=%q want %q", got, sid)
	}
	if phase, _ := el["phase"].(string); phase == "" {
		t.Fatalf("JSON missing phase; el=%v", el)
	}
	ext, ok := el["extension"].(map[string]any)
	if !ok {
		t.Fatalf("JSON missing extension object; el=%v", el)
	}
	if _, ok := ext["connected"].(bool); !ok {
		t.Fatalf("JSON extension.connected missing or not bool; ext=%v", ext)
	}
}```
