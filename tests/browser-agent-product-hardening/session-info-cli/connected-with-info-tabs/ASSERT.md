## Expected

Requirement **B2**:

- HandleCLI nil error; ExitCode 0.
- Stdout ends with `\n`.
- JSON shows extension connected true (nested or top-level).
- Browser tab inventory present: `browser.tabs` (preferred) or nested
  `data.tabs` / documented equivalent with at least one tab URL from the fake
  info job (`example.com` or `shop.example`).
- Prefer that fake extension observed an `info` job (WSJobReceived /
  ObservedJobType=info) — strong signal implementer enqueued type=info.

## Side Effects

- Temp BaseDir cleaned by harness.

## Errors

- Control-only stdout without tabs while connected fails this leaf.

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
	if resp.DispatchTimedOut {
		t.Fatal("session info timed out (connected + info tabs)")
	}
	if resp.CLIErr != "" {
		t.Fatalf("session info should succeed when connected; CLIErr=%q stderr=%q stdout=%q",
			resp.CLIErr, resp.Stderr, resp.Stdout)
	}
	assertExitZero(t, resp)
	assertPrintedTrailingNewline(t, resp.Stdout)

	if !stdoutLooksLikeJSONObject(resp.Stdout) {
		t.Fatalf("stdout should be JSON object; got %s", truncate(resp.Stdout, 300))
	}

	low := strings.ToLower(resp.Stdout)
	// Connected true
	connectedOK := resp.ExtensionConnected ||
		strings.Contains(low, `"connected": true`) ||
		strings.Contains(low, `"connected":true`)
	if !connectedOK {
		t.Fatalf("extension.connected must be true after WS hello; stdout=%s",
			truncate(resp.Stdout, 500))
	}

	// Tabs from info job must appear in stdout
	hasTabURL := strings.Contains(resp.Stdout, "example.com") ||
		strings.Contains(resp.Stdout, "shop.example")
	hasTabsKey := strings.Contains(low, `"tabs"`) || resp.HasTabsKey
	if !hasTabsKey || !hasTabURL {
		t.Fatalf("connected session info must include browser tabs from info job; HasTabsKey=%v TabCount=%d stdout=%s",
			resp.HasTabsKey, resp.TabCount, truncate(resp.Stdout, 600))
	}
	if resp.TabCount == 0 && hasTabsKey {
		// Parsed count may be 0 if shape unexpected; URL presence already checked.
	}

	// Strong preference: info job observed on WS
	if resp.ObservedJobType != "" && resp.ObservedJobType != "info" {
		t.Fatalf("ObservedJobType=%q, want info (or empty if job not observable); JobsSeen=%v",
			resp.ObservedJobType, resp.JobsSeen)
	}
	// If we saw any job, it should be info; if we saw none, still require tabs in stdout
	// (implementer might use a different observation path — tabs in stdout is mandatory).
	if resp.WSJobReceived && resp.ObservedJobType != "info" {
		t.Fatalf("when a job is observed it must be type=info; got %q", resp.ObservedJobType)
	}
}
```
