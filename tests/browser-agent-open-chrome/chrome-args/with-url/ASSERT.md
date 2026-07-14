## Expected

- Nil error.
- ChromeArgs include managed flags and the request URL.

## Side Effects

- Extension extracted under ManagedRoot.

## Errors

- Missing URL in argv fails.

## Exit Code

- N/A.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("chrome-args with-url error: %v", err)
	}
	assertManagedChromeArgsContract(t, resp.ChromeArgs, resp.Layout.DataDir, resp.ExtensionPath, req.URL)
}```
