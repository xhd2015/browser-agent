## Expected

- `popup.js` found under `Chrome-Ext-Browser-Agent`.
- Popup reads **last-known** status via `chrome.storage.session|local.get` and/or
  `chrome.runtime.sendMessage` with a status/getStatus/popupStatus type.
- Health-only `fetch` without storage or SW status query does **not** satisfy.

## Side Effects

- None (read-only FS).

## Errors

- Without last-known, every popup open shows blank/checking phases until network
  and SW respond — worse UX after SW restarts or slow health.

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
	if !resp.FileExists || strings.TrimSpace(resp.PopupJS) == "" {
		t.Fatalf("popup.js missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	js := resp.PopupJS
	if !hasLastKnownStatusPaint(js) {
		t.Fatalf("popup.js must read last-known status (chrome.storage.session|local.get and/or runtime.sendMessage status) before/parallel with live refresh; health-only fetch is insufficient; js=%s",
			truncate(js, 900))
	}
}
```
