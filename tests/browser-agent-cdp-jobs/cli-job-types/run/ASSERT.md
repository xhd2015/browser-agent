## Expected

Requirement **B2**:

- Fake WS observes job type `run`.
- Params include `source` (preferred) or `expression` containing
  `doctest-run-marker` and/or `hello-from-run`.
- CLIErr empty; ExitCode 0; stdout ends with `\n`.

## Side Effects

- Temp script written under BaseDir (harness).

## Errors

- Posting `eval` instead of `run`, or omitting file body, is a failure.

## Exit Code

- 0.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.DispatchTimedOut {
		t.Fatal("run job-type timed out")
	}
	if resp.CLIErr != "" {
		t.Fatalf("run should succeed; CLIErr=%q stderr=%q stdout=%q",
			resp.CLIErr, resp.Stderr, resp.Stdout)
	}
	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)
	assertObservedJobType(t, resp, "run")

	src := paramString(resp.ObservedJobParams, "source", "expression", "expr", "code", "script")
	blob := src + "\n" + resp.ObservedJobRaw
	if !strings.Contains(blob, "doctest-run-marker") && !strings.Contains(blob, "hello-from-run") {
		t.Fatalf("run params missing file content marker; params=%v raw=%s",
			resp.ObservedJobParams, truncate(resp.ObservedJobRaw, 500))
	}
}
```
