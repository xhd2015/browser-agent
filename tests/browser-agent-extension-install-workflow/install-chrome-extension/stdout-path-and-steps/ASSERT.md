## Expected

- CLI returns nil error (`CLIErr` empty).
- Stdout contains `extensions/browser-agent/`.
- Stdout mentions `install` steps: `chrome://extensions`, `Load unpacked`, `Developer mode`.

## Side Effects

- Canonical extension dir created under `TestHome`.

## Errors

- Legacy `{baseDir}/extension/` path in stdout fails.

## Exit Code

- 0 on success.

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
		t.Fatalf("CLI error: %s", resp.CLIErr)
	}
	assertContainsFold(t, resp.Stdout, "extensions/browser-agent/", "chrome://extensions", "load unpacked", "developer mode")
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatal("stdout must end with newline")
	}
}
```