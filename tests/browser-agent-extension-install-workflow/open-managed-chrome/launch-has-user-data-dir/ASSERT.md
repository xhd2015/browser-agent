## Expected

- CLI succeeds.
- `LaunchCallCount == 1`.
- `LaunchArgs` includes `--user-data-dir`.

## Side Effects

- Managed data dir created under root.

## Errors

- Missing user-data-dir fails.

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
	if resp.CLIErr != "" {
		t.Fatalf("CLI error: %s", resp.CLIErr)
	}
	if resp.LaunchCallCount != 1 {
		t.Fatalf("LaunchFn call count=%d, want 1", resp.LaunchCallCount)
	}
	assertArgsHasUserDataDir(t, resp.LaunchArgs)
}
```