## Expected

Requirement **A2**:

- Serve succeeds.
- `meta.json` exists under session dir and parses as JSON.
- Fields:
  - `session_id` matches live session
  - `product` = `browser-agent`
  - `base_url` and/or `session_url` present; session_url contains `/go` + session id when set
- Preferred extras (soft if missing is OK for GREEN? — assert preferred when present):
  - `system_prompt_path` ends with `SYSTEM.md` when present
  - `extension_install_path` non-empty when present
  - `addr` or control_port when present

This leaf **requires** session_id, product, and at least one of base_url/session_url.

## Side Effects

- Only under BaseDir sessions/.

## Errors

- Missing meta blocks CLI discovery.

## Exit Code

- Not asserted.

```go
import (
	"os"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	sid := req.SessionID
	if sid == "" {
		sid = resp.RealSessionID
	}
	if resp.MetaPath == "" {
		t.Fatal("MetaPath empty")
	}
	if _, err := os.Stat(resp.MetaPath); err != nil {
		t.Fatalf("meta.json missing at %s: %v", resp.MetaPath, err)
	}
	assertMetaCore(t, resp, sid)

	// Preferred: system_prompt_path
	if p := stringField(resp.Meta, "system_prompt_path", "systemPromptPath"); p != "" {
		if !strings.HasSuffix(p, "SYSTEM.md") && !strings.HasSuffix(p, "system.md") {
			t.Fatalf("system_prompt_path should end with SYSTEM.md; got %q", p)
		}
	}
}
```
