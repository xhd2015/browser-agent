---
name: browser-agent
description: >-
  Drive a live Chrome session via the browser-agent control plane: start with
  session new (auto-ensures daemon), open the session page, then use nested
  session side commands (session info, create-tab, eval, run, logs, screenshot,
  cdp) against BROWSER_AGENT_SESSION_ID on control port 43761.
---

# Browser Agent Skill

Agent playbook for **`session new` → session page → nested session side commands**
on product **browser-agent** (control port **43761**).

## When to use

- User wants browser automation / inspection through browser-agent
- Need `session eval`, `session run`, `session logs`, `session screenshot`,
  `session create-tab`, or raw `session cdp` against a live tab
- Session is already running and `BROWSER_AGENT_SESSION_ID` is known

## Prerequisites

1. **`browser-agent` on PATH**
2. Session id via `--session-id` or env **`BROWSER_AGENT_SESSION_ID`**
3. Chrome extension loaded (Load unpacked) and connected to the session page

Control server on **127.0.0.1:43761** is started automatically by `session new`
when needed — **agents must not run `browser-agent serve` as the bootstrap step.**

## Agent workflow

### 1. Bootstrap session (agents: this only)

**Do this:**

```bash
browser-agent session new
# or with a fixed id:
# browser-agent session new --session-id <id>
```

`session new` **ensures the daemon**, creates the session, and opens Chrome
(unless `--no-open-chrome`). No separate `serve` process in another terminal.

```bash
export BROWSER_AGENT_SESSION_ID=<id-from-session-new>
```

**Do not** (agent anti-patterns):

```bash
# WRONG for agents — blocks the terminal; session new already starts the daemon
browser-agent serve
browser-agent serve --kill-existing
```

`serve` is only for **humans** who want a long-lived daemon terminal. Agents
should never start it as step 1.

**Optional status probe (read-only; does not start a session):**

```bash
browser-agent serve --status
```

**Deprecated:** `browser-agent serve --session-id <id>` — use `session new` instead.

Open / keep the session page (`/go?session=<id>`) so the extension can attach.
**Keep the session page open** — do not close the tab or navigate that tab away
from `/go?session=`.

### 2. Resolve session and tabs

```bash
export BROWSER_AGENT_SESSION_ID=<id>
browser-agent session info --session-id "$BROWSER_AGENT_SESSION_ID"
```

**session info** returns extension match **and** the open-tab list (`id`, `title`,
`url`, `active`). Always check which tab is **active** before eval/screenshot.

### 3. Side commands (nested under `session`; require session)

```bash
browser-agent session create-tab --session-id "$BROWSER_AGENT_SESSION_ID"
browser-agent session create-tab --session-id "$BROWSER_AGENT_SESSION_ID" https://example.com
browser-agent session eval --session-id "$BROWSER_AGENT_SESSION_ID" '1+1'
browser-agent session run --session-id "$BROWSER_AGENT_SESSION_ID" script.js
browser-agent session logs --session-id "$BROWSER_AGENT_SESSION_ID"
browser-agent session screenshot --session-id "$BROWSER_AGENT_SESSION_ID" -o out.png
browser-agent session cdp --session-id "$BROWSER_AGENT_SESSION_ID" Page.navigate '{"url":"https://example.com"}'
browser-agent session info --session-id "$BROWSER_AGENT_SESSION_ID"
```

Default control base: `http://127.0.0.1:43761`.

**create-tab** opens a tab in the session window (job type `create_tab`). Result
identity is **`tab_id` only** (optional url/window_id). Use that `tab_id` with
`--tab-id` on subsequent jobs.

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
- Prefer **session create-tab** / **session info** for tab lifecycle over inventing a
  raw CDP target graph.

### session cdp + Target.* polyfill

Transport is `chrome.debugger` on **one tab**, not full browser CDP.

| Use | Notes |
|-----|--------|
| `Runtime.evaluate`, `Page.navigate`, `Page.captureScreenshot`, DOM/CSS/Network as needed | Page-scoped CDP via debugger |
| **`Target.*` (polyfilled)** | Tab lifecycle via **chrome.tabs** in the session window; results use **`tab_id`** only |

**Target.* is polyfilled** for create/close/activate/list/info. Soft methods
(setDiscoverTargets, setAutoAttach, attach/detach) are no-op or map to debugger
attach. Other Target methods return a product **polyfill unsupported** error
(not Chrome `-32000 Not allowed`).

Do not expect a real multi-target CDP graph or worker targets. Prefer:

```bash
browser-agent session create-tab --session-id "$BROWSER_AGENT_SESSION_ID" [url]
browser-agent session info --session-id "$BROWSER_AGENT_SESSION_ID"
```

### 4. Extension install

```bash
browser-agent install-chrome-extension
# then chrome://extensions → Developer mode → Load unpacked
```

## CLI reference

```bash
# Agents start here:
browser-agent session new [--session-id]

browser-agent session info|create-tab|eval|run|logs|screenshot|cdp [flags]
browser-agent serve --status          # read-only probe only
browser-agent serve [flags]           # human long-lived daemon — not agent bootstrap
```

Control port **43761**. Session env: **BROWSER_AGENT_SESSION_ID**.
