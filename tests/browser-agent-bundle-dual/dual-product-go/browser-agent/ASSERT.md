## Expected

Requirement **C1**:

- Product id and/or CLI name is `browser-agent`.
- Control port is **43761**.
- Features include `browser-agent`.
- Must not report port 43759 for this product.
- Distinct id (not equal to browser-trace).

## Side Effects

- None (pure constants).

## Errors

- Wrong product id/port breaks SPA + extension host alignment.

## Exit Code

- Not asserted (pure).

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
	if id != "browser-agent" && cli != "browser-agent" {
		t.Fatalf("ProductBrowserAgent id/cli = %q / %q, want browser-agent",
			resp.ProductID, resp.ProductCLIName)
	}
	port := resp.ProductControlPort
	portStr := resp.ProductPortStr
	if port != 43761 && portStr != "43761" {
		t.Fatalf("control port = %d (%q), want 43761", port, portStr)
	}
	if port == 43759 || portStr == "43759" {
		t.Fatal("browser-agent must not use browser-trace port 43759")
	}
	found := false
	for _, f := range resp.ProductFeatures {
		if strings.EqualFold(f, "browser-agent") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("features must include browser-agent; got %v", resp.ProductFeatures)
	}
	if strings.EqualFold(resp.ProductID, "browser-trace") {
		t.Fatal("ProductBrowserAgent id must be distinct from browser-trace")
	}
}
```
