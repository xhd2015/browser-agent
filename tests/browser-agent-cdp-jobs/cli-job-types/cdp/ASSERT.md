## Expected

Requirement **B5**:

- Fake WS observes job type `cdp`.
- Params (or raw envelope) include method `Page.navigate`.
- Preferred: nested url `https://example.com` present somewhere in params/raw.
- CLIErr empty; ExitCode 0; stdout ends with `\n`.

## Side Effects

- None.

## Errors

- Wrong type or missing method is a failure.

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
		t.Fatal("cdp job-type timed out")
	}
	if resp.CLIErr != "" {
		t.Fatalf("cdp should succeed; CLIErr=%q stderr=%q stdout=%q",
			resp.CLIErr, resp.Stderr, resp.Stdout)
	}
	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)
	assertObservedJobType(t, resp, "cdp")

	method := paramString(resp.ObservedJobParams, "method", "cdp_method", "cdpMethod")
	blob := method + "\n" + resp.ObservedJobRaw
	if !strings.Contains(blob, "Page.navigate") {
		t.Fatalf("cdp params missing method Page.navigate; params=%v raw=%s",
			resp.ObservedJobParams, truncate(resp.ObservedJobRaw, 500))
	}
	// Soft: URL preferred but not strictly required if method present.
	_ = strings.Contains(blob, "example.com")
}
```
