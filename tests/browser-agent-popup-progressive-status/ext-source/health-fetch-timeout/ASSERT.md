## Expected

- `popup.js` found under `Chrome-Ext-Browser-Agent`.
- Health/status fetch uses **AbortController** (or `AbortSignal.timeout` /
  Promise.race timeout) so the check cannot hang forever.
- Bare `fetch(.../v1/health)` without signal/abort does **not** satisfy.

## Side Effects

- None (read-only FS).

## Errors

- Without timeout, a stalled control server leaves the popup on infinite
  `checking…` and hides progressive phase failures.

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
	if !hasHealthFetchTimeout(js) {
		t.Fatalf("popup.js health/status fetch must use AbortController/timeout abort (not bare fetch); js=%s",
			truncate(js, 900))
	}
}
```
