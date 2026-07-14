## Expected

- Respawn never healthy → EnsureDaemon error (Q12)

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
	if err == nil && resp.CLIErr == "" {
		t.Fatal("expected error when respawn unhealthy")
	}
}
```
