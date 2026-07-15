## Expected

Requirement **B2**:

- Fake WS observes job type **`create_tab`**.
- Params (or raw envelope) include the URL marker `create-tab-marker` / `example.com`.
- CLIErr empty; ExitCode 0; stdout ends with `\n`.

## Side Effects

- One job pushed over WS.

## Errors

- Wrong type or missing URL in params is a failure.

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
		t.Fatal("create-tab with-url timed out")
	}
	if resp.CLIErr != "" {
		t.Fatalf("create-tab with url should succeed; CLIErr=%q stderr=%q stdout=%q",
			resp.CLIErr, resp.Stderr, resp.Stdout)
	}
	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)
	assertObservedJobType(t, resp, "create_tab")

	urlVal := paramString(resp.ObservedJobParams, "url", "URL", "href")
	blob := urlVal + "\n" + resp.ObservedJobRaw
	if !strings.Contains(blob, "create-tab-marker") && !strings.Contains(blob, "example.com") {
		t.Fatalf("create_tab params missing url; params=%v raw=%s",
			resp.ObservedJobParams, truncate(resp.ObservedJobRaw, 500))
	}
}
```
