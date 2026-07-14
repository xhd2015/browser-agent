## Expected

- Pre-seeded session dir exists.
- `Create("disk-leftover")` returns `errors.Is(err, ErrSessionExists)`.
- Session is not registered (Get returns false).

## Side Effects

- Pre-existing dir unchanged; no meta.json from failed Create.

## Errors

- Create succeeds or wrong error type fails this leaf.

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
	if !browseragent.SessionDirExists(req.BaseDir, req.SessionID) {
		t.Fatal("pre-seeded session dir missing")
	}
	assertErrSessionExists(t, resp.CreateErr)
	reg := browseragent.NewSessionRegistry(req.BaseDir, req.Addr)
	if _, ok := reg.Get(req.SessionID); ok {
		t.Fatal("Get after failed Create should be false")
	}
	assertExitZero(t, resp)
}
```