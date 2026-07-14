## Expected

After implementer lands DELETE /v1/session (**RED** on current code):

- `DeleteStatusCode` is **409**.
- Response body mentions extension connected.

## Side Effects

- Session not deleted (dir and registry remain).

## Errors

- 200/204/405 or missing extension-connected message fails.

## Exit Code

- Not asserted (HTTP leaf).

```go
import (
	"net/http"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.SessionID == "" {
		t.Fatal("harness did not create session")
	}

	if resp.DeleteStatusCode != http.StatusConflict {
		t.Fatalf("DELETE status=%d want 409; body=%s",
			resp.DeleteStatusCode, truncate(resp.DeleteBodyString, 400))
	}
	assertContainsFold(t, resp.DeleteBodyString, "extension", "connected")
}
```