## Expected

- `DefaultAddr` is `127.0.0.1:43761`
- serve `--help` documents `--host` and `--port` (not `--addr` alone)

## Side Effects

- See leaf scenario (may mutate daemon meta, session dirs, or stderr).

## Errors

- Wrong version/port/upgrade/stop behavior fails the assertion.

## Exit Code

- Not asserted unless noted in Expected.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp.Addr != "127.0.0.1:43761" {
		t.Fatalf("DefaultAddr=%q want 127.0.0.1:43761", resp.Addr)
	}
	low := strings.ToLower(resp.HelpText)
	if !strings.Contains(low, "--host") || !strings.Contains(low, "--port") {
		t.Fatalf("serve --help missing --host/--port; help=%s", truncate(resp.HelpText, 600))
	}
	if strings.Contains(low, "--addr") {
		t.Fatal("serve --help still documents deprecated --addr")
	}
}
```
