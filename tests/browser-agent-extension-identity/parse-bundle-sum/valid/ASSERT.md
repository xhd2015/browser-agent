## Expected

Requirement **A1**:

- `ParseBundleSumJS` succeeds (`ParseOK`).
- `Version` is `1.0.1`.
- `MD5` is `a1b2c3d4e5f6789012345678abcdef01` (lowercase hex accepted).
- ExitCode 0.

## Side Effects

- None (pure).

## Errors

- Any parse error fails this leaf.

## Exit Code

- 0.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.ParseOK {
		t.Fatalf("expected ParseOK; ParseErr=%q", resp.ParseErr)
	}
	if resp.Version != "1.0.1" {
		t.Fatalf("Version=%q, want 1.0.1", resp.Version)
	}
	got := strings.ToLower(strings.TrimSpace(resp.MD5))
	want := "a1b2c3d4e5f6789012345678abcdef01"
	if got != want {
		t.Fatalf("MD5=%q, want %q", resp.MD5, want)
	}
	assertExitZero(t, resp)
}
```
