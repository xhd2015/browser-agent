---
name: browser-agent
description: >-
  Drive a live Chrome session via the browser-agent control plane: start serve
  or session new, open the session page, then use nested session side commands
  (session info, eval, run, logs, screenshot, cdp) against BROWSER_AGENT_SESSION_ID
  on control port 43761.
---

# Browser Agent Skill

Agent playbook for **serve or session new → session page → nested session side commands**
on product **browser-agent** (control port **43761**).

## When to use

- User wants browser automation / inspection through browser-agent
- Need `session eval`, `session run`, `session logs`, `session screenshot`, or raw `session cdp` against a live tab
- Session is already running and `BROWSER_AGENT_SESSION_ID` is known

## Prerequisites

1. **`browser-agent` on PATH**
2. Control server listening on **127.0.0.1:43761** (default)
3. Session id via `--session-id` or env **`BROWSER_AGENT_SESSION_ID`**
4. Chrome extension loaded (Load unpacked) and connected to the session page

## Agent workflow

### 1. Bootstrap control server + session

**Option A — dedicated daemon terminal (recommended for long sessions):**

```bash
# terminal 1 — blocking multi-session daemon host
browser-agent serve
# optional:
# browser-agent serve --addr 127.0.0.1:43761 --kill-existing
```

**Option B — one-shot session bootstrap (auto-spawns daemon if needed):**

```bash
browser-agent session new
# or with a fixed id:
# browser-agent session new --session-id <id>
```

**Status probe (read-only, any terminal):**

```bash
browser-agent serve --status
```

**Deprecated:** `browser-agent serve --session-id <id>` still works but is deprecated;
prefer plain `browser-agent serve` (daemon host) or `browser-agent session new`.

Open the session page (`/go?session=<id>`) so the extension can attach.
**Keep the session page open** so the extension stays connected — do not close the tab
or navigate to a different session in the same window.

### 2. Resolve session and tabs

Prefer flag, else env:

```bash
export BROWSER_AGENT_SESSION_ID=<id>
browser-agent session info --session-id "$BROWSER_AGENT_SESSION_ID"
```

**session info** returns extension match **and** the open-tab list (`id`, `title`,
`url`, `active`). Always check which tab is **active** before eval/screenshot.

### 3. Side commands (nested under `session`; require session)

```bash
browser-agent session eval --session-id "$BROWSER_AGENT_SESSION_ID" '1+1'
browser-agent session run --session-id "$BROWSER_AGENT_SESSION_ID" script.js
browser-agent session logs --session-id "$BROWSER_AGENT_SESSION_ID"
browser-agent session screenshot --session-id "$BROWSER_AGENT_SESSION_ID" -o out.png
browser-agent session cdp --session-id "$BROWSER_AGENT_SESSION_ID" Page.navigate '{"url":"https://example.com"}'
browser-agent session info --session-id "$BROWSER_AGENT_SESSION_ID"
```

Default control base: `http://127.0.0.1:43761`.

### Tab targeting (critical)

- **Default:** eval / run / logs / screenshot / cdp target the **active capturable tab**
  in the session window (where `/go?session=<id>` is open).
- **Prefer `--tab-id`:** `browser-agent session eval --tab-id <chromeTabId> '...'` for stable
  targeting on background tabs without focusing them.
- **`--tab-index <n>`** (1-based capturable tabs in session window) is available but unstable;
  stderr warns to prefer `--tab-id`. Mutually exclusive with `--tab-id`.
- Run **`session info`** first (human table or `--json`): tabs with Idx/ID/Role/active,
  `job_target`, and `recommended_cli`.
- Do **not** navigate the session control page away from `/go?session=` — that disconnects
  the extension.
- List tabs with **session info** only — do **not** use CDP `Target.getTargets`.

### session cdp limits (Chrome extension debugger)

Transport is `chrome.debugger` on **one tab**, not full browser CDP.

| Use | Do not use |
|-----|------------|
| `Runtime.evaluate`, `Page.navigate`, `Page.captureScreenshot`, DOM/CSS/Network as needed | **`Target.*`** (`getTargets`, `activateTarget`, `createTarget`, …) |

Forbidden Target methods return:

```text
cdp job failed: {"code":-32000,"message":"Not allowed"}
```

Never:

```bash
browser-agent session cdp Target.getTargets '{}'
browser-agent session cdp Target.activateTarget '{"targetId":"…"}'
```

### 4. Extension install

```bash
browser-agent install-chrome-extension
# then chrome://extensions → Developer mode → Load unpacked
```

## CLI reference

```bash
browser-agent serve [flags]
browser-agent serve --status
browser-agent session new [--session-id]
browser-agent session info|eval|run|logs|screenshot|cdp [flags]
```

Control port **43761**. Session env: **BROWSER_AGENT_SESSION_ID**.