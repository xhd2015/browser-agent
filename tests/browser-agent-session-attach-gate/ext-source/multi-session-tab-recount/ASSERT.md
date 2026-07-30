## Expected

- `background.js` found under `Chrome-Ext-Browser-Agent`.
- Leave path (`tabs.onRemoved` and/or navigate-away in `tabs.onUpdated`) **re-queries**
  open `/go?session=<id>` control tabs (or calls a named recount helper).
- Decision to unregister/detach is based on **remaining count**, not only
  `entry.tabId === closedTabId`.
- Hello telemetry (`collectSessionPageTelemetry`) alone does **not** satisfy this leaf.

## Side Effects

- None (read-only FS).

## Errors

- Single-`entry.tabId` leave handling wrongly tears down the session when a second
  control tab still exists.

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
		t.Fatalf("shell background missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	text := resp.CombinedText
	if !hasMultiSessionTabRecount(text) {
		t.Fatalf("background leave path must re-query remaining /go?session= control tabs (not only entry.tabId); text=%s",
			truncate(text, 900))
	}
}
```
