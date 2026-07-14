## Expected

Requirement **A2**:

- `ParseBundleSumJS` returns a non-nil error (`ParseOK` false).
- `ParseErr` / `ErrText` non-empty.
- ExitCode non-zero.

Also smoke-check empty input separately in-assert via a second call only if
exported; primary path is the leaf fixture.

## Side Effects

- None (pure).

## Errors

- Success (`ParseOK`) fails this leaf.

## Exit Code

- Non-zero.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/browser-agent/browseragent"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ParseOK {
		t.Fatalf("expected parse error; got Version=%q MD5=%q", resp.Version, resp.MD5)
	}
	if strings.TrimSpace(resp.ParseErr) == "" && strings.TrimSpace(resp.ErrText) == "" {
		t.Fatal("ParseErr/ErrText empty on invalid input")
	}
	if resp.ExitCode == 0 {
		t.Fatal("ExitCode=0 on parse error")
	}

	// Empty input must also error (same public contract).
	if _, e := browseragent.ParseBundleSumJS(nil); e == nil {
		t.Fatal("ParseBundleSumJS(nil) should error")
	}
	if _, e := browseragent.ParseBundleSumJS([]byte{}); e == nil {
		t.Fatal("ParseBundleSumJS(empty) should error")
	}
}
```
