## Expected

- `HandleCLI install-firefox-extension --color` returns **nil** (exit **0**).
- Stdout still contains path marker `extensions/browser-agent-firefox/`.
- Stdout contains ANSI: green (`\x1b[32m`) for success head **and/or** yellow
  (`\x1b[33m`) for `warning:` restart note (at least one required; prefer both).

## Side Effects

- Extension extracted under TestHome.

## Errors

- Missing ANSI color or non-zero exit fails.

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
		t.Fatalf("CLIErr = %q, want empty with --color", resp.CLIErr)
	}
	assertContainsFold(t, resp.Stdout, "extensions/browser-agent-firefox/")
	// Extract must land under this leaf's TestHome (HOME isolation).
	if req.TestHome != "" && !strings.Contains(resp.Stdout, req.TestHome) {
		t.Fatalf("stdout path should be under TestHome %q; stdout=%q",
			req.TestHome, truncate(resp.Stdout, 600))
	}
	if !resp.HasANSI {
		t.Fatalf("expected ANSI color in stdout with --color; stdout=%q",
			truncate(resp.Stdout, 600))
	}
	if !resp.HasGreenANSI && !resp.HasYellowANSI {
		t.Fatalf("expected green (\\x1b[32m) success and/or yellow (\\x1b[33m) warning; stdout=%q",
			truncate(resp.Stdout, 600))
	}
}
```
