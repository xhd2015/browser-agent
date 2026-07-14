## Expected

- `WSHelloOK` true.
- Session A (`sess-p4-a`): `ExtensionConnectedA` true, `SupportsA` true.
- Session B (`sess-p4-b`): `ExtensionConnectedB` false.

## Side Effects

- Only A's snapshot reflects hello; B unchanged.

## Errors

- B connected after hello on A only fails the leaf.

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
	if !resp.ExtensionConnectedA {
		t.Fatalf("session A not connected after hello on A; probe=%s", resp.SessionAProbeURL)
	}
	if !resp.SupportsA {
		t.Fatal("session A supports_browser_agent=false after hello on A")
	}
	if resp.ExtensionConnectedB {
		t.Fatalf("session B connected=true after hello only on A; probe=%s", resp.SessionBProbeURL)
	}
}
```