## Expected

- Nil error from Run.
- FirefoxArgs does **not** contain `--load-extension` or `--user-data-dir`.
- No `http://` / `https://` argument (empty URL → no navigation URL).

## Side Effects

- None (pure function).

## Errors

- Presence of managed flags or a URL arg fails.

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
	// Empty slice is allowed; if present, still no managed flags / URLs.
	assertNoManagedFirefoxFlags(t, resp.FirefoxArgs)
	for _, a := range resp.FirefoxArgs {
		if strings.HasPrefix(a, "http://") || strings.HasPrefix(a, "https://") {
			t.Fatalf("empty-url must not include navigation URL; found %q in %v", a, resp.FirefoxArgs)
		}
	}
}
```
