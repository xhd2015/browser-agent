## Expected Output

Stdout exact path + newline.

```
---
version: 2
__PATH__: type=string, example=/tmp/browser-trace-base/2026-07-11-12-00-00-abc, session directory path
---
__PATH__
```

## Expected

- Exit code 0.
- Stdout is exactly `{SessionDir}\n`.
- Stderr mentions **hello** and/or **version** (verbose detail).
- Info milestones may still appear (verbose is additive); at least one of
  listen/session/ready/recording is acceptable but not required beyond hello.

## Side Effects

- Optional log file may include the same verbose lines.

## Errors

- None.

## Exit Code

- 0.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitZero(t, resp)
	assertStdoutSessionPathOnly(t, resp)

	low := strings.ToLower(resp.Stderr)
	hasHello := strings.Contains(low, "hello")
	hasVersion := strings.Contains(low, "version") ||
		strings.Contains(low, "test-mock") ||
		strings.Contains(low, "1.0.0")
	if !hasHello && !hasVersion {
		t.Fatalf("Verbose stderr should mention hello and/or version; stderr:\n%s", resp.Stderr)
	}
}
```
