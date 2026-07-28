## Expected

- Install returns nil error (`CLIErr` empty).
- Stdout keeps P1 markers: `extensions/browser-agent-firefox/`, `about:debugging`,
  Load Temporary.
- **Step 3** opens the **folder** (path) and must **not** select `manifest.json`
  on the same step line.
- **Step 4** selects the file `manifest.json` (not the folder).
- Stdout cautions against **about:addons** (Install from File expects `.xpi`).
- Restart note uses a `warning:` prefix.
- Stdout does **not** suggest `chrome://extensions`.
- Stdout ends with trailing `\n`.

## Side Effects

- Canonical Firefox extension dir created under TestHome.

## Errors

- Conflated single step 3 (“select this folder's manifest.json”), missing step 4,
  missing about:addons/.xpi caution, or missing `warning:` fails.

## Exit Code

- 0 on success.

```go
import (
	"regexp"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.CLIErr != "" {
		t.Fatalf("CLI error: %s", resp.CLIErr)
	}
	out := resp.Stdout
	if out == "" {
		t.Fatal("stdout is empty")
	}
	if !strings.HasSuffix(out, "\n") {
		t.Fatal("stdout must end with newline")
	}

	// P1 markers still required.
	assertContainsFold(t, out,
		"extensions/browser-agent-firefox/",
		"about:debugging",
		"load temporary",
	)
	assertNotContainsFold(t, out, "chrome://extensions")

	// Numbered step lines (tolerate leading whitespace).
	step3 := regexp.MustCompile(`(?mi)^\s*3\.\s*.+$`).FindString(out)
	step4 := regexp.MustCompile(`(?mi)^\s*4\.\s*.+$`).FindString(out)
	if step3 == "" {
		t.Fatalf("expected numbered step 3 (open folder); stdout:\n%s", truncate(out, 900))
	}
	if step4 == "" {
		t.Fatalf("expected numbered step 4 (select manifest.json); stdout:\n%s", truncate(out, 900))
	}
	low3 := strings.ToLower(step3)
	low4 := strings.ToLower(step4)
	if !strings.Contains(low3, "folder") {
		t.Fatalf("step 3 should open the folder; got %q", step3)
	}
	// Step 3 must not also be the "select manifest.json" instruction.
	if strings.Contains(low3, "manifest.json") {
		t.Fatalf("step 3 must open the folder only — select manifest.json is step 4; got %q", step3)
	}
	if !strings.Contains(low4, "manifest.json") {
		t.Fatalf("step 4 should select manifest.json; got %q", step4)
	}
	// Order: step 3 text before step 4 text in the full stdout.
	if strings.Index(out, step3) > strings.Index(out, step4) {
		t.Fatalf("step 3 must appear before step 4; step3=%q step4=%q", step3, step4)
	}

	// about:addons caution + .xpi (Install from File is the wrong surface).
	assertContainsFold(t, out, "about:addons")
	assertContainsFold(t, out, ".xpi")

	// Restart note uses warning: prefix (color leaf checks yellow ANSI separately).
	if !strings.Contains(out, "warning:") {
		t.Fatalf("restart note must use \"warning:\" prefix; stdout:\n%s", truncate(out, 900))
	}
}
```
