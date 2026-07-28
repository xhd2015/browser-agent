## Expected

- `HandleCLI install-firefox-extension --no-color` returns **nil** (exit **0**).
- Stdout contains path marker `extensions/browser-agent-firefox/`.
- Stdout contains **no** ANSI escape sequences (`\x1b`).

## Side Effects

- Extension extracted under TestHome.

## Errors

- ANSI present in stdout or non-zero exit fails.
- Unknown-flag rejection of `--no-color` fails (flag must be accepted).

## Exit Code

- **0**.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode != 0 {
		t.Fatalf("HandleCLI exit = %d, want 0; cliErr=%q stderr=%q",
			resp.ExitCode, resp.CLIErr, truncate(resp.Stderr, 400))
	}
	if resp.CLIErr != "" {
		t.Fatalf("CLIErr = %q, want empty with --no-color (flag must be accepted)", resp.CLIErr)
	}
	assertContainsFold(t, resp.Stdout, "extensions/browser-agent-firefox/")
	if req.TestHome != "" && !strings.Contains(resp.Stdout, req.TestHome) {
		t.Fatalf("stdout path should be under TestHome %q; stdout=%q",
			req.TestHome, truncate(resp.Stdout, 600))
	}
	if resp.HasANSI {
		t.Fatalf("expected no ANSI with --no-color; stdout=%q", truncate(resp.Stdout, 600))
	}
}
```
