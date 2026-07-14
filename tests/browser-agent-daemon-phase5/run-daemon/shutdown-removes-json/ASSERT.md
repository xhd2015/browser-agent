## Expected

- After ctx cancel and RunDaemon exit, `server.json` is **absent**.

## Side Effects

- Removes `{BaseDir}/server.json` on clean shutdown.

## Errors

- File still present after shutdown fails.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.DaemonMetaExists {
		t.Fatalf("server.json still exists at %s after shutdown", resp.DaemonMetaPath)
	}
}
```