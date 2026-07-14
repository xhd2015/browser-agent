## Expected

- `ClientVersion()` returns non-empty string
- Matches `VERSION.txt` at module root

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
	if resp.ClientVersion == "" {
		t.Fatal("ClientVersion() is empty")
	}
	if resp.EmbeddedVer == "" {
		t.Fatal("VERSION.txt missing or empty at module root")
	}
	if resp.ClientVersion != resp.EmbeddedVer {
		t.Fatalf("ClientVersion=%q VERSION.txt=%q", resp.ClientVersion, resp.EmbeddedVer)
	}
}
```
