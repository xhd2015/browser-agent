## Expected

- CLI returns nil error.
- `LaunchCallCount == 1`.
- Stdout mentions extension/profile markers.

## Side Effects

- Managed extension synced.

## Errors

- Unknown command or zero launch calls fail.

## Exit Code

- 0 on success.

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
	assertContainsFold(t, resp.Stdout, "extension", "profile")
}
```