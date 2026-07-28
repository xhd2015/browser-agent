## Expected

- Install returns nil error (`CLIErr` empty).
- Stdout contains `extensions/browser-agent-firefox/`.
- Stdout mentions about:debugging (or `about:debugging`).
- Stdout mentions Load Temporary (Add-on) install steps.
- Stdout does **not** suggest Chrome Load unpacked as the primary path
  (`chrome://extensions` must not appear).
- Stdout ends with trailing `\n`.

## Side Effects

- Canonical Firefox extension dir created under TestHome.

## Errors

- Missing about:debugging markers or Chrome-only instructions fails.

## Exit Code

- 0 on success.

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
	if resp.CLIErr != "" {
		t.Fatalf("CLI error: %s", resp.CLIErr)
	}
	assertContainsFold(t, resp.Stdout,
		"extensions/browser-agent-firefox/",
		"about:debugging",
		"load temporary",
	)
	// Firefox path, not Chrome managed layout.
	assertNotContainsFold(t, resp.Stdout, "chrome://extensions", "managed-chrome")
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatal("stdout must end with newline")
	}
}
```
