## Expected

- `OpenManagedChrome` returns nil error.
- `LaunchCallCount == 1`.
- `LaunchArgs` satisfy managed chrome contract (blank window).
- `OpenResult` populated with layout, extension path, ChromeArgs.

## Side Effects

- Extension synced under ManagedRoot.
- LaunchFn called exactly once.

## Errors

- Zero or multiple LaunchFn calls fails.

## Exit Code

- N/A.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("OpenManagedChrome error: %v", err)
	}
	if resp.LaunchCallCount != 1 {
		t.Fatalf("LaunchCallCount = %d, want 1", resp.LaunchCallCount)
	}
	if resp.OpenResult == nil {
		t.Fatal("OpenResult is nil")
	}
	dataDir := resp.OpenResult.Layout.DataDir
	extPath := resp.OpenResult.ExtensionPath
	assertManagedChromeArgsContract(t, resp.LaunchArgs, dataDir, extPath, "")
	if len(resp.OpenResult.ChromeArgs) > 0 {
		assertManagedChromeArgsContract(t, resp.OpenResult.ChromeArgs, dataDir, extPath, "")
	}
}```
