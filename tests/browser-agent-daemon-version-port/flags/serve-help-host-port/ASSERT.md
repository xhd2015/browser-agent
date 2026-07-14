## Expected

- `serve --help` documents `--host` and `--port` (not `--addr`)

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
	low := strings.ToLower(resp.HelpText)
	if !strings.Contains(low, "--host") || !strings.Contains(low, "--port") {
		t.Fatalf("serve help missing host/port flags: %s", truncate(resp.HelpText, 500))
	}
	if strings.Contains(low, "--addr") {
		t.Fatal("serve help still shows --addr")
	}
}
```
