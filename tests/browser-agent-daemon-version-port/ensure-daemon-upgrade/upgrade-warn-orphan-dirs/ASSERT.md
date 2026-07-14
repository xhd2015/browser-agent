## Expected

- Disconnected orphan session → stderr lists id; session dir removed

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
	assertContainsFold(t, resp.Stderr, "orphan", req.OrphanID)
	if resp.OrphanDirExists {
		t.Fatalf("orphan dir still exists for %s", req.OrphanID)
	}
}
```
