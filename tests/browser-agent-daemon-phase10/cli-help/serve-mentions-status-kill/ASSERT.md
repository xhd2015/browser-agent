## Expected

- `HandleCLI serve --help` returns **nil** (`CLIErr` empty).
- Help text contains **`--status`** and read-only hint (`read-only` or `read only`).
- Help text contains **`--kill-existing`**.
- Help text contains **`--stop`**.
- Help text contains **deprecation** note for **`serve --session-id`** (or
  `serve flags` section marks `--session-id` deprecated).

## Side Effects

- Read-only.

## Errors

- Missing serve operator flags or deprecation note fails.

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
		t.Fatalf("serve --help should return nil error; got %q", resp.CLIErr)
	}
	text := strings.ToLower(resp.HelpText)
	assertContainsFold(t, resp.HelpText, "--status", "--kill-existing", "--stop")
	if !strings.Contains(text, "read-only") && !strings.Contains(text, "read only") {
		t.Fatalf("serve help must document --status as read-only; got:\n%s",
			truncate(resp.HelpText, 900))
	}
	deprecationOK := strings.Contains(text, "deprecat") &&
		(strings.Contains(text, "serve --session-id") ||
			strings.Contains(text, "--session-id") && strings.Contains(text, "deprecated"))
	if !deprecationOK {
		t.Fatalf("serve help must deprecate serve --session-id; got:\n%s",
			truncate(resp.HelpText, 900))
	}
}
```