package browseragent

import "fmt"

// FormatSystemPrompt builds the agent SYSTEM.md playbook for a session.
// sessionID is accepted for API stability but MUST NOT be embedded in the body;
// nested CLI resolves the control id via --session-id or BROWSER_AGENT_SESSION_ID.
func FormatSystemPrompt(sessionID string) string {
	_ = sessionID // keep signature; do not embed concrete control id
	// Note: body is a raw string; avoid backtick characters inside.
	return fmt.Sprintf(`# browser-agent session

You are co-piloting a browser session via the **browser-agent** control plane.

Session id is resolved from flag or env:
- pass --session-id <id>, or
- export BROWSER_AGENT_SESSION_ID=<id>

## Side commands

Use these nested CLI recipes (session id resolved from flag or env):

- browser-agent session info
- browser-agent session eval '/* JS expression */'
- browser-agent session run path/to/script.js
- browser-agent session logs
- browser-agent session screenshot

Optional escape hatch:

- browser-agent session cdp <method> [json]

## Tab targeting (critical)

- **Default:** eval / run / logs / screenshot / cdp target the **active capturable tab**
  in the session window (the Chrome window where /go?session= is open).
- **Prefer --tab-id:** pass an explicit Chrome tab id for stable targeting:
  browser-agent session eval --tab-id <id> '...'
- **--tab-index** (1-based capturable tab list) is available but unstable; prefer --tab-id.
- Run **session info** first: human table or --json lists tabs (Idx, ID, Role, active)
  plus job_target and recommended_cli.
- If the user's content is in a **background** tab, use --tab-id from session info —
  do **not** navigate the session control page away from /go?session= (disconnects extension).
- Do **not** try to switch tabs via CDP Target.* (see below).

## session cdp limits (Chrome extension debugger)

Jobs use chrome.debugger attached to **one tab**. That is a restricted CDP
transport — **not** full remote debugging.

**Allowed (typical page-scoped methods):**
- Runtime.* (e.g. Runtime.evaluate)
- Page.* (e.g. Page.navigate, Page.captureScreenshot)
- DOM.*, CSS.*, Network.* (when needed)

**Forbidden / will fail with code -32000 message "Not allowed":**
- **Target.*** — especially Target.getTargets, Target.activateTarget,
  Target.createTarget, Target.closeTarget, Target.attachToTarget
- Browser-level / multi-target control over sendCommand

Never call:
- browser-agent session cdp Target.getTargets …
- browser-agent session cdp Target.activateTarget …

To list tabs, use **session info** only (chrome.tabs), not Target.getTargets.

## Notes

- Control server default: %s (port %s)
- Jobs are enqueue-and-wait over HTTP POST /v1/jobs
- The Chrome extension executes jobs over WebSocket /v1/ws
- Prefer session info → eval/screenshot on the correct active tab before inventing raw CDP
`, DefaultAddr, DefaultControlPortString())
}
