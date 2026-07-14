---
name: browser-agent
description: >-
  Drive a live Chrome session via the browser-agent control plane: start serve,
  open the session page, then use nested session side commands (session info,
  eval, run, logs, screenshot, cdp) against BROWSER_AGENT_SESSION_ID on control
  port 43761.
---

# Browser Agent Skill

Agent playbook for **serve → session page → nested session side commands** on product
**browser-agent** (control port **43761**).

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

### 1. Start control server (if not running)

```bash
browser-agent serve --session-id <id>
# optional:
# browser-agent serve --addr 127.0.0.1:43761 --no-open-chrome
```

Open the session page (`/go?session=<id>`) so the extension can attach.

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

### Active tab (critical)

- **eval / run / logs / screenshot / cdp always target the active tab** in the
  last-focused window. There is no `--tab-id` flag yet.
- If the page the user cares about is a **background** tab:
  - Ask the user to focus that tab, then re-run `session info` + eval; or
  - `session cdp Page.navigate '{"url":"…"}'` on the **active** tab (navigates
    that tab; does not switch Chrome's selected tab).
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
browser-agent session info|eval|run|logs|screenshot|cdp [flags]
browser-agent skill --show
browser-agent skill --list
browser-agent skill --install …
```

Control port **43761**. Session env: **BROWSER_AGENT_SESSION_ID**.
