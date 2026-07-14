## Expected

- `HandleCLI --help` returns **nil** (`CLIErr` empty).
- Full help text contains **`session new`**.
- Full help describes **`serve`** as a blocking daemon host (`blocking` or `daemon host`).
- When `CaptureBriefUsage`, brief usage (bare invoke stdout) contains **`session new`**.

## Side Effects

- Read-only.

## Errors

- Missing `session new` or blocking serve description fails.

## Exit Code

- 0 for `--help`.

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
		t.Fatalf("--help should return nil error; got %q", resp.CLIErr)
	}
	full := strings.ToLower(resp.HelpText)
	if !strings.Contains(full, "session new") {
		t.Fatalf("full help must document session new; got:\n%s", truncate(resp.HelpText, 900))
	}
	if !strings.Contains(full, "blocking") && !strings.Contains(full, "daemon host") {
		t.Fatalf("full help must describe serve as blocking daemon host; got:\n%s",
			truncate(resp.HelpText, 900))
	}
	if req.CaptureBriefUsage {
		brief := strings.ToLower(resp.BriefUsageText)
		if !strings.Contains(brief, "session new") {
			t.Fatalf("briefUsage must mention session new; got:\n%s", truncate(resp.BriefUsageText, 600))
		}
	}
}
```