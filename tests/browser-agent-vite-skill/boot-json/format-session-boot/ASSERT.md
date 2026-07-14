## Expected

Requirement **D1**:

- `FormatSessionBootJSON` returns parseable JSON object.
- `session_id` equals request BootSessionID (`boot-sess-fixed`).
- `product` is `browser-agent`.
- `control_port` is `43761` (number or string).

## Side Effects

- None.

## Errors

- Invalid JSON / wrong fields fails.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.BootParseOK {
		t.Fatalf("FormatSessionBootJSON not valid JSON: %q err=%q",
			resp.BootJSON, resp.ErrText)
	}
	wantSID := req.BootSessionID
	if wantSID == "" {
		wantSID = "boot-sess-fixed"
	}
	if resp.BootSessionID != wantSID {
		t.Fatalf("session_id = %q, want %q; raw=%s",
			resp.BootSessionID, wantSID, resp.BootJSON)
	}
	if resp.BootProduct != "browser-agent" {
		t.Fatalf("product = %q, want browser-agent; raw=%s",
			resp.BootProduct, resp.BootJSON)
	}
	if resp.BootControlPort != 43761 && resp.BootPortStr != "43761" {
		t.Fatalf("control_port = %d (%q), want 43761; raw=%s",
			resp.BootControlPort, resp.BootPortStr, resp.BootJSON)
	}
}
```
