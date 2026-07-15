## Expected

- `HandleCLI(["assets","ensure"])` returns **nil** error.
- After ensure: `CacheCompleteSP` true and `CacheCompleteExt` true for
  browser-agent at current version (`v` + ClientVersion / `v0.2.0`).

## Side Effects

- Cache trees under isolated XDG only (session-page + extension).

## Errors

- Non-nil CLI error or either kind incomplete fails.

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
	if !resp.CacheCompleteSP {
		t.Fatal("CacheComplete(session-page)=false after assets ensure")
	}
	if !resp.CacheCompleteExt {
		t.Fatal("CacheComplete(extension)=false after assets ensure")
	}
	assertExitZero(t, resp)
}
```
