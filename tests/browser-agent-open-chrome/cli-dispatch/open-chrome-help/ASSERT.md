## Expected

- `HandleCLI open-managed-chrome --help` returns nil error.
- Printed help ends with trailing `\n`.
- Help mentions `open-managed-chrome` command name.
- Help mentions `managed` profile (or `managed-chrome`).
- Help documents `--root` flag.

## Side Effects

- None.

## Errors

- Non-nil error on help fails.

## Exit Code

- 0.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("open-managed-chrome --help error: %v", err)
	}
	if resp.CLIErr != "" {
		t.Fatalf("CLIErr = %q, want empty on --help", resp.CLIErr)
	}
	text := resp.Stdout + resp.Stderr
	if text == "" {
		t.Fatal("help output empty")
	}
	if !strings.HasSuffix(text, "\n") {
		t.Fatal("help must end with trailing newline")
	}
	low := strings.ToLower(text)
	if !strings.Contains(low, "open-managed-chrome") {
		t.Fatalf("help must mention open-managed-chrome; got:\n%s", truncate(text, 800))
	}
	if !strings.Contains(low, "managed") {
		t.Fatalf("help must mention managed profile; got:\n%s", truncate(text, 800))
	}
	if !strings.Contains(low, "--root") {
		t.Fatalf("help must document --root flag; got:\n%s", truncate(text, 800))
	}
}
```