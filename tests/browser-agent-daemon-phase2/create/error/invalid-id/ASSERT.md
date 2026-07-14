## Expected

- `Create("a/b")` returns non-nil error.
- `errors.Is(err, ErrSessionExists)` is false.
- Session directory does not exist on disk.

## Side Effects

- None (no artifacts).

## Errors

- Nil error or ErrSessionExists fails this leaf.

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
	if resp.CreateErr == nil {
		t.Fatal("Create err=nil want validation error")
	}
	if resp.CreateErrIsSessionExists {
		t.Fatalf("errors.Is(err, ErrSessionExists)=true want false; err=%v", resp.CreateErr)
	}
	if browseragent.SessionDirExists(req.BaseDir, req.SessionID) {
		t.Fatal("session dir should not exist after invalid Create")
	}
	assertExitZero(t, resp)
}
```