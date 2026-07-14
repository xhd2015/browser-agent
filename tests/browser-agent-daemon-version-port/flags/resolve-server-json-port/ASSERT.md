## Expected

- `ResolveControlBaseURLWithHostPort` with empty host/port reads `server.json` addr

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
	want := "http://" + resp.Addr
	if resp.ResolveBaseURL != want {
		t.Fatalf("resolved=%q want %q", resp.ResolveBaseURL, want)
	}
}
```
