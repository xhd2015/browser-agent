## Expected

- `WSHelloOK` true.
- Session B (`sess-p4-b`): `ExtensionConnectedB` true, `SupportsB` true.
- Session A (`sess-p4-a`): `ExtensionConnectedA` false.

## Side Effects

- Only B's snapshot reflects hello; A unchanged.

## Errors

- A connected after hello on B only fails the leaf.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.WSHelloOK {
		t.Fatal("WSHelloOK=false")
	}
	if !resp.ExtensionConnectedB {
		t.Fatalf("session B not connected after hello on B; probe=%s", resp.SessionBProbeURL)
	}
	if !resp.SupportsB {
		t.Fatal("session B supports_browser_agent=false after hello on B")
	}
	if resp.ExtensionConnectedA {
		t.Fatalf("session A connected=true after hello only on B; probe=%s", resp.SessionAProbeURL)
	}
}
```