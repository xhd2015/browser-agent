## Expected

- `Exists("absent-id")` returns false on fresh registry with no disk dir.

## Side Effects

- None.

## Errors

- Exists true fails this leaf.

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
	if browseragent.SessionDirExists(req.BaseDir, req.ExistsSessionID) {
		t.Fatal("session dir should not exist")
	}
	if resp.Exists {
		t.Fatalf("Exists(%q)=true want false", req.ExistsSessionID)
	}
	assertExitZero(t, resp)
}
```