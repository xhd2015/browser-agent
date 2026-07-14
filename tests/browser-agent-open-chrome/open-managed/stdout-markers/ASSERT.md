## Expected

- `HandleCLI` returns nil error (`CLIErr` empty).
- Stdout non-empty; ends with trailing `\n`.
- Stdout mentions managed profile (substring `managed`).
- Stdout mentions `profile` or `data` and extension path or version.
- `LaunchCallCount == 1` via test hook.

## Side Effects

- Extension synced; LaunchFn invoked once.

## Errors

- Non-nil CLI error or missing stdout markers fails.

## Exit Code

- 0.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("HandleCLI open-managed-chrome error: %v", err)
	}
	if resp.CLIErr != "" {
		t.Fatalf("CLIErr = %q, want empty", resp.CLIErr)
	}
	if resp.LaunchCallCount != 1 {
		t.Fatalf("LaunchCallCount = %d, want 1", resp.LaunchCallCount)
	}
	assertStdoutTrailingNewline(t, resp.Stdout)
	assertContainsFold(t, resp.Stdout, "managed")
	if !containsAnyFold(resp.Stdout, "profile", "data") {
		t.Fatalf("stdout missing profile/data marker; got:\n%s", truncate(resp.Stdout, 800))
	}
	if !containsAnyFold(resp.Stdout, "extension", "load-extension") {
		t.Fatalf("stdout missing extension marker; got:\n%s", truncate(resp.Stdout, 800))
	}
}

func containsAnyFold(haystack string, needles ...string) bool {
	low := strings.ToLower(haystack)
	for _, n := range needles {
		if strings.Contains(low, strings.ToLower(n)) {
			return true
		}
	}
	return false
}```
