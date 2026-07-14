## Expected

- Health JSON includes `product`, `daemon_version`, `base_dir`

## Side Effects

- See leaf scenario (may mutate daemon meta, session dirs, or stderr).

## Errors

- Wrong version/port/upgrade/stop behavior fails the assertion.

## Exit Code

- Not asserted unless noted in Expected.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp.HealthJSON == nil {
		t.Fatal("health JSON nil")
	}
	for _, k := range []string{"product", "daemon_version", "base_dir"} {
		if _, ok := resp.HealthJSON[k]; !ok {
			t.Fatalf("health missing field %q; got %v", k, resp.HealthJSON)
		}
	}
	if resp.HealthJSON["product"] != "browser-agent" {
		t.Fatalf("product=%v", resp.HealthJSON["product"])
	}
}
```
