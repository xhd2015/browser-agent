## Expected

- CLIErr empty; exit 0
- Help lists summary, paths, entries, show

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp.DispatchTimedOut {
		t.Fatal("timed out")
	}
	if resp.CLIErr != "" {
		t.Fatalf("CLIErr=%q", resp.CLIErr)
	}
	assertExitZero(t, resp)
	text := strings.ToLower(combinedCLIText(resp))
	for _, want := range []string{"summary", "paths", "entries", "show"} {
		if !strings.Contains(text, want) {
			t.Fatalf("help missing %q; got:\n%s", want, truncate(combinedCLIText(resp), 800))
		}
	}
}
```
