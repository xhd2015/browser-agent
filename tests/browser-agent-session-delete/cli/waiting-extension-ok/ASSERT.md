## Expected

After implementer lands session delete (**RED** on current code):

- `HandleCLI` returns nil; `ExitCode` 0.
- Combined stdout/stderr mentions `deleted` and the session id.
- `SessionDirExists` false for the deleted id.
- `SessionInList` false (`GET /v1/sessions` does not include id).

## Side Effects

- Session directory removed from `{BaseDir}/sessions/{id}/`.
- Registry entry removed.

## Errors

- Non-zero exit, missing deleted message, dir still present, or id still listed fails.

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
		t.Fatal("session delete timed out")
	}
	if resp.SessionID == "" {
		t.Fatal("harness did not create session")
	}

	assertExitZero(t, resp)
	out := combinedOutput(resp)
	assertContainsFold(t, out, "deleted", resp.SessionID)

	if resp.SessionDirExists {
		t.Fatalf("session dir still exists after delete: %s",
			browseragent.SessionDirPath(req.BaseDir, resp.SessionID))
	}
	if resp.SessionInList {
		t.Fatalf("session still in GET /v1/sessions after delete; list=%s",
			truncate(resp.SessionsListRaw, 500))
	}
}
```