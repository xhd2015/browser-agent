## Expected

Requirement **C2**:

- Product id and/or CLI name is `browser-trace`.
- Control port is **43759**.
- Features include `browser-trace`.
- Must not report port 43761 for this product.

## Side Effects

- None.

## Errors

- Swapping to agent port collides with browser-agent control plane.

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
		t.Fatalf("control port = %d (%q), want 43759", port, portStr)
	}
	if port == 43761 || portStr == "43761" {
		t.Fatal("browser-trace must not use browser-agent port 43761")
	}
	found := false
	for _, f := range resp.ProductFeatures {
		if strings.EqualFold(f, "browser-trace") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("features must include browser-trace; got %v", resp.ProductFeatures)
	}
}
```
