## Expected

- `background.js` found under `Chrome-Ext-Browser-Agent`.
- Boot path (`chrome.runtime.onInstalled`, `onStartup`, and/or a named heal
  helper wired from boot) **queries tabs** (`chrome.tabs.query` or equivalent)
  and **parses** `/go?session=` (`parseGoSessionFromURL` / `isSessionGoPageURL`
  / `/go?session=` filter).
- Iterating empty `sessions.keys()` alone does **not** satisfy this leaf.

## Side Effects

- None (read-only FS).

## Errors

- Without tab rediscover, cold SW restart leaves open `/go` control tabs unbound
  — no session windowId, no WS, no attach until the page reloads or re-registers.

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
	if !hasBootRediscoverGoTabs(text) {
		t.Fatalf("background boot path must chrome.tabs.query open /go?session= tabs (onInstalled/onStartup/heal helper); sessions.keys() alone is insufficient; text=%s",
			truncate(text, 900))
	}
}
```
