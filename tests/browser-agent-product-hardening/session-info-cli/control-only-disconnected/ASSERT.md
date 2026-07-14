## Expected

Requirement **B1**:

- HandleCLI nil error; ExitCode 0 preferred.
- Stdout ends with `\n`.
- Stdout is JSON-ish object including session id and connected=false (or nested
  `extension.connected` false).
- Phase waiting* when present (not extension_connected).
- **No fabricated page tabs** (no `tabs` array of http(s) page URLs).
- Browser unavailable is indicated (`browser` null, `browser_error`, or clear
  message that tab inventory needs a connected extension).

## Side Effects

- Temp BaseDir cleaned by harness.

## Errors

- Inventing tabs while disconnected is a product bug.

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
		t.Fatal("session info timed out (disconnected control-only)")
	}
	if resp.CLIErr != "" {
		t.Fatalf("session info should succeed when session exists; CLIErr=%q stderr=%q stdout=%q",
			resp.CLIErr, resp.Stderr, resp.Stdout)
	}
	assertExitZero(t, resp)
	assertPrintedTrailingNewline(t, resp.Stdout)

	wantID := resp.RealSessionID
	if wantID == "" {
		wantID = req.SessionID
	}
	if wantID != "" && !strings.Contains(resp.Stdout, wantID) {
		t.Fatalf("stdout missing session id %q; stdout=%s", wantID, truncate(resp.Stdout, 400))
	}
	if !stdoutLooksLikeJSONObject(resp.Stdout) {
		t.Fatalf("stdout should be JSON object; got %s", truncate(resp.Stdout, 300))
	}

	// Connected must be false (nested or top-level).
	if resp.ExtensionConnected {
		t.Fatal("extension.connected must be false without WS hello")
	}
	low := strings.ToLower(resp.Stdout)
	if !strings.Contains(low, "connected") {
		t.Fatalf("stdout must mention connected field; stdout=%s", truncate(resp.Stdout, 400))
	}
	// Reject true when we parsed false already; also reject obvious true without parse.
	if strings.Contains(low, `"connected": true`) || strings.Contains(low, `"connected":true`) {
		t.Fatalf("connected must be false without extension; stdout=%s", truncate(resp.Stdout, 400))
	}

	if resp.Phase != "" {
		pl := strings.ToLower(resp.Phase)
		if strings.Contains(pl, "extension_connected") ||
			(strings.Contains(pl, "connected") && !strings.Contains(pl, "wait")) {
			// allow waiting_extension; reject extension_connected
			if strings.Contains(pl, "extension_connected") {
				t.Fatalf("phase=%q, want waiting* before hello", resp.Phase)
			}
		}
	}

	if fabricatedPageTabs(resp.Stdout, resp.ParsedJSON) {
		t.Fatalf("must not fabricate page tabs when extension disconnected; stdout=%s",
			truncate(resp.Stdout, 500))
	}
	if resp.HasTabsKey && resp.TabCount > 0 {
		t.Fatalf("tabs inventory must be empty/absent when disconnected; TabCount=%d stdout=%s",
			resp.TabCount, truncate(resp.Stdout, 400))
	}

	// browser unavailable signal
	if !resp.BrowserUnavailable {
		// Accept control-only snapshot that still has waiting phase / connected false
		// and an install/hint string — require at least one explicit signal.
		hintOK := strings.Contains(low, "waiting") ||
			strings.Contains(low, "install") ||
			strings.Contains(low, "extension") && strings.Contains(low, "connect")
		if !hintOK {
			t.Fatalf("disconnected info must signal browser unavailable or waiting/install hint; stdout=%s",
				truncate(resp.Stdout, 500))
		}
	}
}
```
