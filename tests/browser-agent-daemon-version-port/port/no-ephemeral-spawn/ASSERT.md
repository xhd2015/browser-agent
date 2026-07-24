## Expected

- `EnsureDaemon` with empty addr uses `127.0.0.1:43761` (not ephemeral `:0`)

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

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if !resp.SpawnFnCalled {
		t.Fatal("SpawnFn not called")
	}
	if resp.SpawnAddrUsed != "127.0.0.1:43761" {
		t.Fatalf("spawn addr intent=%q want 127.0.0.1:43761 (EnsureDaemon resolves empty Addr to DefaultAddr)", resp.SpawnAddrUsed)
	}
	// Actual listen may use a free port so the leaf is host-safe; still must not be ":0" alone.
	if resp.Meta.Addr != "" && (resp.Meta.Addr == ":0" || strings.HasSuffix(resp.Meta.Addr, ":0")) {
		t.Fatalf("meta.Addr=%q must not be ephemeral :0", resp.Meta.Addr)
	}
}
```
