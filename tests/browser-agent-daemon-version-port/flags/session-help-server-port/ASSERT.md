## Expected

- `session new --help` documents `--host` and `--server-port`

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
	if !strings.Contains(low, "--host") || !strings.Contains(low, "--server-port") {
		t.Fatalf("session help missing --host/--server-port: %s", truncate(resp.HelpText, 500))
	}
	if strings.Contains(low, "--addr") {
		t.Fatal("session help still shows --addr")
	}
}
```
