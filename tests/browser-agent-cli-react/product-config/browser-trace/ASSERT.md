## Expected

Requirement **C2**:

- ProductID or CLIName is `browser-trace`.
- Control port is **43759**.
- Must not be 43761 (agent port).

## Side Effects

- None.

## Errors

- Missing dual export fails compile of this leaf until implementer adds
  `ProductBrowserTrace`.

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
	id := strings.ToLower(resp.ProductID)
	cli := strings.ToLower(resp.ProductCLIName)
	if id != "browser-trace" && cli != "browser-trace" {
		t.Fatalf("ProductBrowserTrace id/cli = %q / %q, want browser-trace",
			resp.ProductID, resp.ProductCLIName)
	}
	port := resp.ProductControlPort
	portStr := resp.ProductPortStr
	if port != 43759 && portStr != "43759" {
		t.Fatalf("browser-trace control port = %d (%q), want 43759", port, portStr)
	}
	if port == 43761 || portStr == "43761" {
		t.Fatal("browser-trace must not use browser-agent port 43761")
	}
}
```
