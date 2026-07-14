## Expected

- `SessionNew` path detects foreign listener → non-zero error + hint

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
	if resp.CLIErr == "" {
		t.Fatal("expected SessionNew/EnsureDaemon failure on foreign port")
	}
	if !resp.ForeignHintSeen {
		t.Fatalf("missing foreign hint; stderr=%q err=%q", resp.Stderr, resp.CLIErr)
	}
}
```
