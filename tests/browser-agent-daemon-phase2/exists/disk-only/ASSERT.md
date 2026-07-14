## Expected

- Session dir exists on disk without registry entry.
- `Exists("disk-only")` returns true.

## Side Effects

- Pre-seeded directory only.

## Errors

- Exists false fails this leaf.

## Exit Code

- 0.

```go
import (
	"testing"

	"github.com/xhd2015/browser-agent/browseragent"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !browseragent.SessionDirExists(req.BaseDir, req.ExistsSessionID) {
		t.Fatal("seeded session dir missing")
	}
	if !resp.Exists {
		t.Fatalf("Exists(%q)=false want true for disk-only dir", req.ExistsSessionID)
	}
	assertExitZero(t, resp)
}
```