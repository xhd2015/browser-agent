## Expected

- `background.js` found under `Chrome-Ext-Browser-Agent`.
- Boot heal is **stronger** than `for (const sessionId of sessions.keys())
  connectSession(...)` alone: must **tabs.query rediscover** `/go?session=` and
  **reconnect** discovered sessions.
- Presence of the obsolete keys loop is allowed **alongside** real heal, but
  keys-loop as the only boot behavior must fail this leaf.

## Side Effects

- None (read-only FS).

## Errors

- Shipping only the empty-keys reconnect pattern leaves open control tabs orphaned
  after every SW restart — the core P3 regression.

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
	if !hasHealStrongerThanEmptySessionsKeysLoop(text) {
		t.Fatalf("background boot heal must be stronger than empty sessions.keys() → connectSession loop (require tabs.query rediscover of /go + reconnect); text=%s",
			truncate(text, 900))
	}
}
```
