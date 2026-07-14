## Expected Output

```text
Session __SESSION_ID__                                    Ready
Extension connected (v1.0.0)
  Idx  ID        Active  Role          Title
  1    111               session-page  Browser Agent Session
  2    222       *       user          Example Domain
Job target  idx 2 / tab 222  (active in session window)
Recommended: browser-agent session eval --tab-id 222 '...'
Keep the session page (/go?session=…) open — navigating it away disconnects the extension.
```

## Expected

- `HandleCLI session info` (human default) returns nil; `ExitCode` 0.
- Stdout is **not** JSON-only.
- Stdout lists **both** enriched tab rows (idx 1 session-page, idx 2 user with active `*`).
- Stdout contains **`Job target  idx 2 / tab 222  (active in session window)`** (underscores in reason normalized to spaces).
- Stdout contains **`Recommended:`** line with **`--tab-id 222`**.
- Stdout contains session-page keep-open hint (same as `formatEnrichedTabsTable`).

## Side Effects

- Info job enqueued when extension connected.

## Errors

- Truncated table (only first tab), bare `Job target` label, or missing footer lines fails (**RED** until `formatCompactEnrichedSessionInfo` matches `formatEnrichedTabsTable` footer logic).

## Exit Code

- **0**.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.DispatchTimedOut {
		t.Fatal("session info timed out")
	}
	if resp.CLIErr != "" {
		t.Fatalf("session info should succeed; CLIErr=%q stderr=%q stdout=%q",
			resp.CLIErr, truncate(resp.Stderr, 300), truncate(resp.Stdout, 500))
	}
	if resp.ExitCode != 0 {
		t.Fatalf("ExitCode=%d want 0", resp.ExitCode)
	}

	trimmed := strings.TrimSpace(resp.Stdout)
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		t.Fatalf("default session info must be human table output, not JSON-only; stdout=%s",
			truncate(resp.Stdout, 500))
	}

	// Compact human path must render full enriched table + footer (not tabs[0] only).
	assert.Output(t, resp.Stdout, `---
version: 2
---
...2 lines omitted...
  Idx  ID        Active  Role          Title
  1    111               session-page  Browser Agent Session
  2    222       *       user          Example Domain
Job target  idx 2 / tab 222  (active in session window)
Recommended: browser-agent session eval --tab-id 222 '...'
Keep the session page (/go?session=…) open — navigating it away disconnects the extension.`)
}
```