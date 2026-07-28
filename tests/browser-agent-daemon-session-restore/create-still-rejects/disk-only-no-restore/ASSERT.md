## Expected

- Create(`sess-diskonly`) returns `errors.Is(err, ErrSessionExists)`.
- Session dir still exists on disk.

## Side Effects

- Pre-seeded dir unchanged (no Create artifacts required).

## Errors

- Create succeeds or wrong error fails this leaf.

## Exit Code

- 0.

```go
import (
	"testing"

	"github.com/xhd2015/browser-agent/browseragent"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertErrSessionExists(t, resp.CreateErr)
	if !resp.CreateErrIsSessionExists {
		t.Fatal("CreateErrIsSessionExists=false want true")
	}
	if !browseragent.SessionDirExists(req.BaseDir, "sess-diskonly") {
		t.Fatal("pre-seeded session dir missing after failed Create")
	}
	assertExitZero(t, resp)
}
```
