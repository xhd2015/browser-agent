## Expected

After implementer lands session delete (**RED** on current code):

- `HandleCLI` returns non-nil error; `ExitCode` 1.
- Error text mentions extension connected (or equivalent).
- `SessionDirExists` true — session not deleted.
- `SessionInList` true — session still registered.
- `ExtensionConnectedAfterReject` true.

## Side Effects

- Session directory and registry entry unchanged.

## Errors

- Exit 0, missing extension-connected message, or session removed fails.

## Exit Code

- 1.

```go
import (
	"testing"
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
	if !resp.ExtensionConnectedBeforeDelete {
		t.Fatal("extension should be connected before delete attempt")
	}

	assertExitOne(t, resp)
	out := combinedOutput(resp)
	assertContainsFold(t, out, "extension", "connected")

	if !resp.SessionDirExists {
		t.Fatal("session dir should remain after connected rejection")
	}
	if !resp.SessionInList {
		t.Fatalf("session should remain in registry; list=%s",
			truncate(resp.SessionsListRaw, 500))
	}
	if !resp.ExtensionConnectedAfterReject {
		t.Fatal("extension should still be connected after rejected delete")
	}
}
```