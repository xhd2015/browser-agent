## Expected

- CLI returns non-nil error (`CLIErr` non-empty).
- Error text contains `unknown command` (or equivalent).
- Error/help text mentions `open-managed-chrome` guidance, not successful open-chrome dispatch.
- `LaunchCallCount == 0`.

## Side Effects

- No chrome launch.

## Errors

- Successful open-chrome dispatch fails this leaf.

## Exit Code

- Non-zero.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.CLIErr == "" {
		t.Fatal("expected CLI error for removed open-chrome command")
	}
	low := strings.ToLower(resp.CLIErr)
	if !strings.Contains(low, "unknown command") {
		t.Fatalf("expected unknown command error; got %q", resp.CLIErr)
	}
	if resp.LaunchCallCount != 0 {
		t.Fatalf("LaunchFn should not be called; count=%d", resp.LaunchCallCount)
	}
}
```