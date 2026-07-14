## Expected

- `HandleCLI session info --json` returns nil; `ExitCode` 0.
- Stdout JSON includes **top-level** `tabs` array with at least one entry containing **`index`** field.
- Stdout JSON includes **top-level** **`job_target`** object with **`tab_index`** (value **2** in fixture).
- Top-level `job_target.tab_id` present (value **222** in fixture).
- Nested-only fields under `browser{}` are insufficient (spec requires enriched session info JSON).

## Side Effects

- Info job enqueued when extension connected.

## Errors

- Missing `index` / `job_target.tab_index` fields fails (RED until implementer lands).

## Exit Code

- **0**.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.DispatchTimedOut {
		t.Fatal("session info --json timed out")
	}
	if resp.CLIErr != "" {
		t.Fatalf("session info --json should succeed; CLIErr=%q stderr=%q stdout=%q",
			resp.CLIErr, truncate(resp.Stderr, 300), truncate(resp.Stdout, 400))
	}
	if resp.ExitCode != 0 {
		t.Fatalf("ExitCode=%d want 0", resp.ExitCode)
	}

	// Spec: enriched fields at top level of session info --json (not only browser sub-object).
	if resp.StdoutJSON == nil {
		t.Fatal("failed to parse --json stdout")
	}
	topTabs, hasTopTabs := resp.StdoutJSON["tabs"].([]any)
	if !hasTopTabs {
		t.Fatalf("top-level tabs[] missing from --json stdout (nested browser.tabs is insufficient); stdout=%s",
			truncate(resp.Stdout, 700))
	}
	hasTabIndex := false
	for _, item := range topTabs {
		if m, ok := item.(map[string]any); ok {
			if _, ok := m["index"]; ok {
				hasTabIndex = true
				break
			}
		}
	}
	if !hasTabIndex {
		t.Fatalf("top-level tabs[].index missing from --json stdout; stdout=%s", truncate(resp.Stdout, 700))
	}

	topJobTarget, ok := resp.StdoutJSON["job_target"].(map[string]any)
	if !ok {
		t.Fatalf("top-level job_target missing from --json stdout; stdout=%s", truncate(resp.Stdout, 700))
	}
	resp.JobTargetJSON = topJobTarget
	tabIndex, ok := resp.JobTargetJSON["tab_index"]
	if !ok {
		t.Fatalf("job_target.tab_index missing; job_target=%v stdout=%s",
			resp.JobTargetJSON, truncate(resp.Stdout, 700))
	}
	if jsonNumberToInt64(tabIndex) != 2 {
		t.Fatalf("job_target.tab_index=%v want 2", tabIndex)
	}
	tabID, ok := resp.JobTargetJSON["tab_id"]
	if !ok || jsonNumberToInt64(tabID) != 222 {
		t.Fatalf("job_target.tab_id=%v want 222; job_target=%v", tabID, resp.JobTargetJSON)
	}
}
```