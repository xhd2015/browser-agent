## Expected

Requirement **F2**:

- Default port is **43761**.
- Accept `DefaultAddr` ending in `:43761` and/or `DefaultPort` / `DefaultControlPort` == `43761`.
- Must not be 43759 (browser-trace port).

## Side Effects

- None.

## Errors

- Wrong default port breaks extension host permission + CLI docs.

## Exit Code

- Not asserted.

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
	port := resp.DefaultPort
	if port == "" && resp.DefaultAddr != "" {
		if i := strings.LastIndex(resp.DefaultAddr, ":"); i >= 0 {
			port = resp.DefaultAddr[i+1:]
		}
	}
	if port != "43761" && !strings.HasSuffix(resp.DefaultAddr, ":43761") && resp.DefaultAddr != "43761" {
		t.Fatalf("default control port/addr = port %q addr %q, want 43761", port, resp.DefaultAddr)
	}
	if port == "43759" || strings.HasSuffix(resp.DefaultAddr, ":43759") {
		t.Fatalf("default must not be browser-trace port 43759; addr=%q", resp.DefaultAddr)
	}
}
```
