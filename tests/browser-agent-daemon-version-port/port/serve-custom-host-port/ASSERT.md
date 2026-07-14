## Expected

- `serve --host 127.0.0.1 --port <free>` binds and health OK

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
	if !resp.HealthOK {
		t.Fatalf("daemon not healthy at %s", resp.Addr)
	}
}
```
