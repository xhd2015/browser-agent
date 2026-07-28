## Expected

- `contentScript.js` found under `Firefox-Ext-Browser-Agent`.
- Retains **`__BROWSER_AGENT_EXT__`** marker.
- Calls **`sendMessage`** via `browser.runtime` or `chrome.runtime`.
- Sends **`type: "register"`** (or `type:"register"`) with **`session_id`**.
- Reads session from go page: `URLSearchParams` / `location.search` / `/go`+`session`
  and/or **`data-session-id`**.

## Side Effects

- None (read-only FS).

## Errors

- Marker-only stub without register prevents WS attach on session pages.

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
		t.Fatalf("Firefox contentScript missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	text := resp.CombinedText
	if !strings.Contains(text, "__BROWSER_AGENT_EXT__") {
		t.Fatalf("contentScript must keep __BROWSER_AGENT_EXT__ marker; text=%s", truncate(text, 500))
	}
	low := strings.ToLower(text)
	if !strings.Contains(low, "sendmessage") {
		t.Fatalf("contentScript must call browser/chrome.runtime.sendMessage; text=%s", truncate(text, 500))
	}
	if !strings.Contains(low, "runtime") {
		t.Fatalf("contentScript must use runtime.sendMessage; text=%s", truncate(text, 500))
	}
	if !strings.Contains(low, "register") {
		t.Fatalf("contentScript must send register message type; text=%s", truncate(text, 500))
	}
	if !strings.Contains(low, "session_id") {
		t.Fatalf("contentScript must include session_id in register payload; text=%s", truncate(text, 500))
	}
	hasSessionFromPage := strings.Contains(low, "urlsearchparams") ||
		strings.Contains(low, "location.search") ||
		strings.Contains(low, "data-session-id") ||
		strings.Contains(low, "/go?session=") ||
		(strings.Contains(low, "session") && (strings.Contains(low, "search") || strings.Contains(low, "/go")))
	if !hasSessionFromPage {
		t.Fatalf("contentScript must read session id from go page URL or data-session-id; text=%s",
			truncate(text, 500))
	}
}
```
