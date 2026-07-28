## Expected

- `background.js` found under `Firefox-Ext-Browser-Agent`.
- Source contains job type branch tokens for each of:
  `info`, `eval`, `run`, `logs`, `screenshot`, `create_tab`
  (quoted / `case` / `===` forms preferred; bare token OK for longer names).

## Side Effects

- None (read-only FS).

## Errors

- Missing any core type fails dispatch completeness (Phase-1 stub-only fails).

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertBackgroundPresent(t, req, resp)
	types := []string{"info", "eval", "run", "logs", "screenshot", "create_tab"}
	for _, jt := range types {
		if !jobTypeTokenPresent(src, jt) {
			t.Fatalf("background missing job type branch token %q; path=%v snippet=%s",
				jt, resp.FoundPaths, truncate(src, 600))
		}
	}
}
```
