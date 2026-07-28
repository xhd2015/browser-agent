## Expected

- Help succeeds (nil error or empty CLIErr preferred; non-fatal if help returns nil).
- Stdout (or combined help text) contains `--browser`.
- Mentions both `chrome` and `firefox` as values (case-insensitive).

## Side Effects

- None.

## Errors

- Missing `--browser` documentation fails.

## Exit Code

- 0 preferred.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	// Help may return nil; only fail on unexpected hard errors with empty output.
	if resp == nil {
		t.Fatal("resp is nil")
	}
	text := resp.Stdout
	if text == "" {
		text = resp.Stderr
	}
	if text == "" {
		if err != nil {
			t.Fatalf("help failed with no output: %v", err)
		}
		t.Fatal("help output is empty")
	}
	assertContainsFold(t, text, "--browser")
	low := strings.ToLower(text)
	if !strings.Contains(low, "firefox") {
		t.Fatalf("help should mention firefox as a --browser value; got:\n%s", truncate(text, 800))
	}
	if !strings.Contains(low, "chrome") {
		t.Fatalf("help should mention chrome as a --browser value; got:\n%s", truncate(text, 800))
	}
}
```
