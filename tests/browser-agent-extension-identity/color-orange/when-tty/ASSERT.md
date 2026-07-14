## Expected

Requirement **D1**:

- `ColorOut` contains ESC (`\x1b`).
- `ColorOut` contains `38;5;208` (256-color orange).
- Prefer reset `\x1b[0m`.
- Original input text still present inside the colored string.

## Side Effects

- None (pure).

## Errors

- Missing ESC or 208 fails the leaf.

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
	out := resp.ColorOut
	if out == "" {
		t.Fatal("ColorOut empty")
	}
	if !strings.Contains(out, "\x1b") {
		t.Fatalf("ColorOut missing ESC when isTTY; got %q", out)
	}
	if !strings.Contains(out, "38;5;208") {
		t.Fatalf("ColorOut missing orange SGR 38;5;208; got %q", out)
	}
	if !strings.Contains(out, req.ColorInput) && !strings.Contains(out, "mismatch") {
		t.Fatalf("ColorOut should retain message text; got %q", out)
	}
}
```
