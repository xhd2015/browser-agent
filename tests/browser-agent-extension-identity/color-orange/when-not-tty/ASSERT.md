## Expected

Requirement **D2**:

- `ColorOut` equals input string (or is a plain copy without ANSI).
- `ColorOut` does **not** contain ESC (`\x1b`).

## Side Effects

- None (pure).

## Errors

- Any ESC in output fails the leaf.

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
	if strings.Contains(out, "\x1b") {
		t.Fatalf("ColorOut must not contain ESC when !isTTY; got %q", out)
	}
	if out != req.ColorInput {
		// Accept if implementer trims; still must be plain equal for contract.
		if strings.TrimSpace(out) != strings.TrimSpace(req.ColorInput) {
			t.Fatalf("ColorOut=%q, want plain input %q", out, req.ColorInput)
		}
	}
}
```
