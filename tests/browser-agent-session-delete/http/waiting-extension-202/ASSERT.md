## Expected

After implementer lands DELETE /v1/session (**RED** on current code):

- `DeleteStatusCode` is **200** or **204**.
- `SessionDirExists` false.
- `SessionInList` false.

## Side Effects

- Session directory removed; registry entry removed.

## Errors

- 405 method not allowed, 404, or session still listed fails.

## Exit Code

- Not asserted (HTTP leaf).

```go
import (
	"net/http"
	"testing"

	"github.com/xhd2015/browser-agent/browseragent"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.SessionID == "" {
		t.Fatal("harness did not create session")
	}

	if resp.DeleteStatusCode != http.StatusOK && resp.DeleteStatusCode != http.StatusNoContent {
		t.Fatalf("DELETE status=%d want 200 or 204; body=%s",
			resp.DeleteStatusCode, truncate(resp.DeleteBodyString, 400))
	}
	if resp.SessionDirExists {
		t.Fatalf("session dir still exists after DELETE: %s",
			browseragent.SessionDirPath(req.BaseDir, resp.SessionID))
	}
	if resp.SessionInList {
		t.Fatalf("session still in GET /v1/sessions; list=%s",
			truncate(resp.SessionsListRaw, 500))
	}
}
```