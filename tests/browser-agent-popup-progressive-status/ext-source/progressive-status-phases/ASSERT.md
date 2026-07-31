## Expected

- `popup.html` and/or `popup.js` found under `Chrome-Ext-Browser-Agent`.
- Distinct progressive phase surfaces for **all three**:
  1. daemon / control health
  2. session / WebSocket
  3. debugger / armed / attach
- Accept flexible naming (`status-daemon`, `status-session`, `status-debugger`,
  `data-phase=…`, etc.).
- Only `#ctrl-status` (control health) without session + debugger hooks does
  **not** satisfy.

## Side Effects

- None (read-only FS).

## Errors

- Single-status popup cannot show “daemon up but WS waiting” vs “armed with
  debugger” — operators cannot tell which stage failed.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.FileExists {
		t.Fatalf("popup sources missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	html := resp.PopupHTML
	js := resp.PopupJS
	if strings.TrimSpace(html) == "" && strings.TrimSpace(js) == "" {
		t.Fatalf("popup.html and popup.js empty under ModuleRoot=%s; found=%v",
			req.ModuleRoot, resp.FoundPaths)
	}
	if !hasProgressiveStatusPhases(html, js) {
		t.Fatalf("popup must expose distinct progressive phase hooks for daemon/control, session/ws, and debugger/armed; single ctrl-status is insufficient; html=%s js=%s",
			truncate(html, 500), truncate(js, 400))
	}
}
```
