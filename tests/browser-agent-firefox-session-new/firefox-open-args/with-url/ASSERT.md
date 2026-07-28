## Expected

- Nil error from Run.
- FirefoxArgs contains the request URL (exact element or substring).
- FirefoxArgs does **not** contain `--load-extension` or `--user-data-dir` (any form).

## Side Effects

- None (pure function).

## Errors

- Missing URL or presence of managed flags fails.

## Exit Code

- N/A (package API).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if len(resp.FirefoxArgs) == 0 {
		t.Fatal("FirefoxArgs is empty; want session URL present")
	}
	assertNoManagedFirefoxFlags(t, resp.FirefoxArgs)
	found := false
	for _, a := range resp.FirefoxArgs {
		if a == req.URL || strings.Contains(a, req.URL) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("FirefoxArgs should include URL %q; args=%v", req.URL, resp.FirefoxArgs)
	}
}
```
