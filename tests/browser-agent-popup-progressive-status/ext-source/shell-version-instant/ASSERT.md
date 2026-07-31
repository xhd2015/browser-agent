## Expected

- `popup.html` and/or `popup.js` found under `Chrome-Ext-Browser-Agent`.
- Version/package identity element present (`#pkg-version` or equivalent).
- Version painted from **sync** path (`bundle-sum.js` / `BROWSER_AGENT_BUNDLE_VERSION`),
  not only after health `fetch` resolves.

## Side Effects

- None (read-only FS).

## Errors

- Blocking first paint of package identity on network makes the popup feel hung
  when the control server is down or slow.

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
		t.Fatalf("popup shell sources missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	html := resp.PopupHTML
	js := resp.PopupJS
	if strings.TrimSpace(html) == "" && strings.TrimSpace(js) == "" {
		t.Fatalf("popup.html and popup.js empty under ModuleRoot=%s; found=%v",
			req.ModuleRoot, resp.FoundPaths)
	}
	if !hasInstantShellVersion(html, js) {
		t.Fatalf("popup must paint package version instantly (pkg-version + sync bundle-sum / BROWSER_AGENT_BUNDLE_VERSION); not only after health fetch; html=%s js=%s",
			truncate(html, 400), truncate(js, 400))
	}
}
```
