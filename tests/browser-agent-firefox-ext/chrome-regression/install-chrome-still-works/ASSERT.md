## Expected

- Install returns nil error.
- Stdout contains `extensions/browser-agent/` (Chrome canonical segment).
- Stdout mentions `chrome://extensions` and `Load unpacked`.
- Stdout ends with trailing `\n`.

## Side Effects

- Chrome extension extracted under TestHome managed-chrome layout.

## Errors

- Regression of Chrome install path fails.

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
	assertContainsFold(t, resp.Stdout, "extensions/browser-agent/", "chrome://extensions", "load unpacked")
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatal("stdout must end with newline")
	}
}
```
