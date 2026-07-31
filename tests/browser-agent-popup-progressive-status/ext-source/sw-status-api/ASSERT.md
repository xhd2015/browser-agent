## Expected

- `background.js` found under `Chrome-Ext-Browser-Agent`.
- SW can report status for the popup via **at least one** of:
  - `chrome.runtime.onMessage` handler for type `status` / `getStatus` /
    `popupStatus` (or equivalent) that can return session + debugger info
  - `chrome.storage.session` / `local` write of last-known status including
    session armed and debugger/attach signals
- `register`-only onMessage and control-plane `sendSessionStatus` alone do
  **not** satisfy this leaf.

## Side Effects

- None (read-only FS).

## Errors

- Without a SW status surface, the popup can only probe HTTP health and cannot
  show session/WS or debugger armed state.

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
	if !resp.FileExists || strings.TrimSpace(resp.BackgroundJS) == "" {
		t.Fatalf("background.js missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	bg := resp.BackgroundJS
	if !hasSWStatusAPI(bg) {
		t.Fatalf("background must expose popup status API (onMessage status/getStatus/popupStatus and/or chrome.storage last-known with session+debugger); register-only / WS sendSessionStatus alone is insufficient; text=%s",
			truncate(bg, 900))
	}
}
```
