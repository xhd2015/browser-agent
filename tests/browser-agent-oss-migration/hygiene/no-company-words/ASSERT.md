## Expected

- `rg` for REQUIREMENT-DESIGN masking-table originals returns **no** matches when excluding `.git/`, `node_modules/`, `dist/`.

## Side Effects

- None.

## Errors

- Any match line fails the leaf (first 20 lines shown).

## Exit Code

- rg exit 1 (no matches) is success; exit 0 with output is failure.

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
	if resp.RunErr != "" && resp.ExitCode != 1 {
		t.Fatalf("rg failed: %s\noutput:\n%s", resp.RunErr, resp.Stdout)
	}
	if len(resp.RGMatches) > 0 {
		show := strings.Join(resp.RGMatches, "\n")
		if len(resp.RGMatches) > 20 {
			show = strings.Join(resp.RGMatches[:20], "\n") + "\n…"
		}
		t.Fatalf("company words found (%d matches):\n%s", len(resp.RGMatches), show)
	}
}
```