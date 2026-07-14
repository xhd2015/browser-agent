## Expected

- Disconnected registered session → warn + RemoveAll dir

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
	assertContainsFold(t, resp.Stderr, req.OrphanID)
	if resp.OrphanDirExists {
		t.Fatalf("orphan dir should be removed for %s", req.OrphanID)
	}
}
```
