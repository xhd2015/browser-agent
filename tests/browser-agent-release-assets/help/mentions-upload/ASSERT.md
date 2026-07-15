## Expected Output

Help text (stdout and/or stderr) mentions the opt-in upload flag and ends with a
newline on the primary help stream.

```text
version: 2
# flexible help wording; require --upload token and a trailing newline on stdout if non-empty
```

## Expected

- Exit code **0**.
- Combined stdout+stderr contains the substring **`--upload`** (exact flag form).
- If stdout is non-empty, it ends with trailing `\n`. If help is printed only on
  stderr, stderr ends with trailing `\n`. At least one of stdout/stderr is
  non-empty and ends with `\n`.

## Side Effects

- None (help only; no archives, no network).

## Errors

- Non-zero exit, empty help, missing `--upload`, or missing trailing newline fails.

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
	assertExitZero(t, resp)

	out := combinedOut(resp)
	if strings.TrimSpace(out) == "" {
		t.Fatalf("help output empty; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(out, "--upload") {
		t.Fatalf("help missing --upload; out=%s", truncate(out, 600))
	}

	// Trailing newline on whichever stream carried the help text.
	switch {
	case resp.Stdout != "" && strings.HasSuffix(resp.Stdout, "\n"):
		// ok
	case resp.Stderr != "" && strings.HasSuffix(resp.Stderr, "\n"):
		// ok (some CLIs write usage to stderr)
	default:
		t.Fatalf("help text must end with trailing newline on stdout or stderr; stdout=%q stderr=%q",
			truncate(resp.Stdout, 200), truncate(resp.Stderr, 200))
	}
}
```
