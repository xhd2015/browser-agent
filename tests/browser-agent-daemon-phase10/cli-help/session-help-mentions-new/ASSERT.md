## Expected

- `HandleCLI session --help` returns **nil** (`CLIErr` empty).
- Help contains **`session new`** with operator hint (**`Ensure daemon`** or
  **`create session`**).
- Help documents **`session new`** `--session-id` with **auto-generate** when
  omitted.
- Bare `session` (no subcommand) brief usage contains **`session new`**.

## Side Effects

- Read-only.

## Errors

- Missing session new section or auto-generate hint fails.

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
	if resp.CLIErr != "" {
		t.Fatalf("session --help should return nil error; got %q", resp.CLIErr)
	}
	text := strings.ToLower(resp.HelpText)
	if !strings.Contains(text, "session new") {
		t.Fatalf("session help must document session new; got:\n%s", truncate(resp.HelpText, 900))
	}
	if !strings.Contains(text, "ensure daemon") && !strings.Contains(text, "create session") {
		t.Fatalf("session new help must describe ensure/create flow; got:\n%s",
			truncate(resp.HelpText, 900))
	}
	if !strings.Contains(text, "auto-generate") && !strings.Contains(text, "auto generate") {
		t.Fatalf("session new help must document auto-generate --session-id; got:\n%s",
			truncate(resp.HelpText, 900))
	}
	if req.CaptureBriefUsage {
		brief := strings.ToLower(resp.BriefUsageText)
		if !strings.Contains(brief, "session new") {
			t.Fatalf("bare session briefUsage must mention session new; got:\n%s",
				truncate(resp.BriefUsageText, 600))
		}
	}
}
```