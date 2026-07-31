## Expected

- `background.js` found under `Chrome-Ext-Browser-Agent`.
- Heal / boot rediscover path **registers or binds** discovered sessions and
  calls **`connectSession`** (or `handleRegisterMessage` / `maybeRegisterGoTab`
  which connect) for those session ids.
- Empty-map `for (sessions.keys()) connectSession` alone does **not** satisfy.

## Side Effects

- None (read-only FS).

## Errors

- Rediscover without reconnect leaves the control plane without an extension WS
  for that session — jobs queue with no worker until manual re-register.

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
	if !hasReconnectWSOnHeal(text) {
		t.Fatalf("background heal path must connectSession/register for tabs.query-discovered /go sessions; empty sessions.keys() connect loop is insufficient; text=%s",
			truncate(text, 900))
	}
}
```
