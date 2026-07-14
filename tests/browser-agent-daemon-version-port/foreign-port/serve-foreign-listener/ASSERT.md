## Expected

- `serve` on foreign-occupied port exits non-zero
- Stderr hints control port / `--server-port` / `--port`

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
	if resp.CLIErr == "" && resp.ExitCode == 0 {
		t.Fatal("expected serve failure on foreign port")
	}
	if !resp.ForeignHintSeen {
		t.Fatalf("missing foreign-port hint; stderr=%q err=%q", resp.Stderr, resp.CLIErr)
	}
}
```
