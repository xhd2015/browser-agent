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

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if !resp.SpawnFnCalled {
		t.Fatal("SpawnFn not called")
	}
	if resp.SpawnAddrUsed != "127.0.0.1:43761" {
		t.Fatalf("spawn addr=%q want 127.0.0.1:43761 (no pickEphemeralAddr)", resp.SpawnAddrUsed)
	}
	if resp.Meta.Addr != "" && !strings.HasSuffix(resp.Meta.Addr, ":43761") {
		t.Fatalf("meta.Addr=%q want :43761 suffix", resp.Meta.Addr)
	}
}
```
