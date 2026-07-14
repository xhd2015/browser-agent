## Expected

Requirement **P1**:

- `background.js` found under `Chrome-Ext-Browser-Agent` (public/ or root/src/build).
- Source connects per-session WS — contains `/v1/ws?session=` **or** builds WS URL with
  `?session=` query alongside `/v1/ws`.
- Still references `/v1/ws` or `ws://` control path.

## Side Effects

- None (read-only FS).

## Errors

- Global-only `/v1/ws` socket breaks per-session routing on multi-session daemon.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.FileExists || strings.TrimSpace(resp.CombinedText) == "" {
		t.Fatalf("shell background missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	text := resp.CombinedText
	if !strings.Contains(text, "/v1/ws") && !strings.Contains(text, "ws://") {
		t.Fatalf("background must reference /v1/ws or ws://; text=%s", truncate(text, 500))
	}
	if !hasPerSessionWSURL(text) {
		t.Fatalf("background must use per-session WS URL (/v1/ws?session= or ?session= on WS); text=%s",
			truncate(text, 600))
	}
}
```