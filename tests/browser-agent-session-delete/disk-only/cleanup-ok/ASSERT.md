## Expected

After implementer lands disk-only delete (**RED** on current code):

- `HandleCLI` returns nil; `ExitCode` 0.
- `SessionDirExists` false after delete.
- Combined output mentions deleted or success for the disk-only id.

## Side Effects

- Disk-only directory removed; `server.json` unchanged.

## Errors

- Non-zero exit or dir still present fails.

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
	if resp.DispatchTimedOut {
		t.Fatal("disk-only delete timed out")
	}
	if resp.SessionID == "" {
		t.Fatal("harness did not set disk-only session id")
	}

	assertExitZero(t, resp)
	if resp.SessionDirExists {
		t.Fatalf("disk-only dir still exists: %s",
			browseragent.SessionDirPath(req.BaseDir, resp.SessionID))
	}
}
```