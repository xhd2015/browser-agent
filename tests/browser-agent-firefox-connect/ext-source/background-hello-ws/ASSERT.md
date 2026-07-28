## Expected

- `background.js` found under `Firefox-Ext-Browser-Agent`.
- References **WebSocket** and control path **`/v1/ws`** (or `ws://`).
- Per-session URL: `/v1/ws?session=` or `?session=` combined with `/v1/ws`.
- Sends **hello** (`type: "hello"` or clear hello send path).
- Identifies product: **`browser_product`** and/or **firefox** in hello/telemetry path.

## Side Effects

- None (read-only FS).

## Errors

- Stub without WS hello never attaches to control plane.

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
	if !resp.FileExists || strings.TrimSpace(resp.CombinedText) == "" {
		t.Fatalf("Firefox background missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	text := resp.CombinedText
	low := strings.ToLower(text)

	if !strings.Contains(low, "websocket") {
		t.Fatalf("background must open WebSocket; text=%s", truncate(text, 600))
	}
	if !strings.Contains(text, "/v1/ws") && !strings.Contains(low, "ws://") {
		t.Fatalf("background must reference /v1/ws or ws://; text=%s", truncate(text, 600))
	}
	if !hasPerSessionWSURL(text) {
		t.Fatalf("background must use per-session WS URL (/v1/ws?session= or ?session= on WS); text=%s",
			truncate(text, 600))
	}
	if !strings.Contains(low, "hello") {
		t.Fatalf("background must send hello on connect; text=%s", truncate(text, 600))
	}
	if !hasHelloBrowserProductFirefox(text) {
		t.Fatalf("background hello path must include browser_product and/or firefox; text=%s",
			truncate(text, 600))
	}
}
```
