## Expected

Requirement **F1**:

- HTTP 200 from `GET /v1/session`.
- `session_id` matches live session when present.
- `extension_install_path` (or camelCase) is non-empty.
- Preferred: path is absolute and contains `extension` segment under BaseDir.

## Side Effects

- Extract may populate `{BaseDir}/extension/{version}` during serve.

## Errors

- Empty install path blocks install guidance / chrome open args.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	assertHTTPStatus(t, resp, http.StatusOK)
	sid := req.SessionID
	if sid == "" {
		sid = resp.RealSessionID
	}
	if resp.SessionIDField != "" && sid != "" && resp.SessionIDField != sid {
		t.Fatalf("session_id=%q, want %q", resp.SessionIDField, sid)
	}
	p := resp.SessionJSONExtensionInstallPath
	if strings.TrimSpace(p) == "" {
		// Also accept path only in meta.json if implementer delays session JSON field —
		// but F1 prefers GET /v1/session. Fail hard on empty.
		t.Fatalf("extension_install_path missing/empty in /v1/session body=%s",
			truncate(resp.BodyString, 500))
	}
	if !filepath.IsAbs(p) {
		t.Fatalf("extension_install_path should be absolute; got %q", p)
	}
	if !strings.Contains(p, "extension") {
		t.Fatalf("extension_install_path should contain extension segment; got %q", p)
	}
}
```
