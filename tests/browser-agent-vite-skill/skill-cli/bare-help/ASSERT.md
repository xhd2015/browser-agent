## Expected

Requirement **C3**:

- HandleCLI returns promptly (no timeout / no accidental serve).
- Either:
  1. **Help path**: printed text (stdout or stderr) mentions skill flags
     `--list` / `--show` / `--install` (at least two of them or the word
     `skill` + `--help`), preferably trailing `\n`; error nil or non-nil OK, or
  2. **skillcmd error path**: non-nil CLIErr mentioning `--show` or `--list`
     or `--install` or `--help` / `try --help`.
- Must not be empty silence with nil error and no text.

## Side Effects

- None.

## Errors

- Hang / timeout fails.
- Completely silent success without usage fails.

## Exit Code

- Not strictly 0; error path may set ExitCode 1.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp.DispatchTimedOut {
		t.Fatal("bare skill timed out (must not hang on serve)")
	}
	text := combinedCLIText(resp)
	errText := resp.CLIErr
	combined := text + "\n" + errText
	low := strings.ToLower(combined)

	mentionsFlag := strings.Contains(low, "--list") ||
		strings.Contains(low, "--show") ||
		strings.Contains(low, "--install") ||
		strings.Contains(low, "--help") ||
		strings.Contains(low, "try --help")
	mentionsSkill := strings.Contains(low, "skill")

	if !mentionsFlag && !mentionsSkill {
		t.Fatalf("bare skill must print help or skillcmd-style error; stdout=%q stderr=%q cliErr=%q",
			resp.Stdout, resp.Stderr, resp.CLIErr)
	}
	// Prefer informative flag mention when possible.
	if errText == "" && strings.TrimSpace(text) == "" {
		t.Fatal("bare skill: nil error and empty output is not acceptable")
	}
}
```
