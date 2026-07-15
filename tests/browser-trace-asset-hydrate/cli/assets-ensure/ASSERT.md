## Expected

- `HandleCLI(["assets","ensure"])` returns **nil** error.
- After ensure: `CacheCompleteExt` true for product `browser-trace` at current
  version (`v` + ClientVersion / `v0.2.0`).

## Side Effects

- Extension cache tree under isolated XDG only (`…/browser-trace/…/extension`).

## Errors

- Non-nil CLI error or extension incomplete fails.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.CLIErr != nil {
		t.Fatalf("HandleCLI assets ensure err=%v stdout=%q stderr=%q",
			resp.CLIErr, truncate(resp.CLIStdout, 300), truncate(resp.CLIStderr, 400))
	}
	if !resp.CacheCompleteExt {
		t.Fatal("CacheComplete(browser-trace, extension)=false after assets ensure")
	}
	assertExitZero(t, resp)
}
```
