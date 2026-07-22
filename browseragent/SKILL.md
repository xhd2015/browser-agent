---
name: browser-agent
description: >-
  Use when research a browser or web page, or a non-api url mentioned. Drive a live Chrome session: start with
  session new (auto-ensures daemon), open the session page, then use nested
  session side commands (session info, create-tab, eval, run, logs, screenshot,
  cdp).
---

# Browser Agent Skill

Drive **browser-agent** (control **127.0.0.1:43761**): **`session new` once** → keep session page → side commands with the **same** `--session-id`.

Requires: `browser-agent` on PATH; extension connected to the session page.

## Flow

### 1. `session new` (once per task)

```bash
browser-agent session new
```

Example stdout (ids differ):

```text
session-id: sess-xqbsmo

Session URL: http://127.0.0.1:43761/go?session=sess-xqbsmo
Control:     http://127.0.0.1:43761
…
```

Reuse the **`session-id:`** value (here `sess-xqbsmo`) as `--session-id` on every later command. Do **not** call `session new` again to refresh the id. New pages → **`session create-tab`**, not another `session new`.

```bash
# WRONG — two sessions
browser-agent session new
browser-agent session new
```

Second `session new` only if the first session is gone/unusable after cleanup, or the user wants isolation.

**Keep** the session page on `/go?session=<id>` — do not close it or navigate that tab away.

### 2. Example commands (same session id)

Assume `session-id: sess-xqbsmo` and a content tab `216774025` from `session info` / `create-tab`.

```bash
# Inspect session + tabs (do this first)
browser-agent session info --session-id sess-xqbsmo
browser-agent session info --session-id sess-xqbsmo --json

# Open tabs in the session window
browser-agent session create-tab --session-id sess-xqbsmo
browser-agent session create-tab --session-id sess-xqbsmo https://example.com
browser-agent session create-tab --session-id sess-xqbsmo 'https://example.com/path?q=1'

# Eval JS (prefer --tab-id for background tabs)
browser-agent session eval --session-id sess-xqbsmo 'document.title'
browser-agent session eval --session-id sess-xqbsmo --tab-id 216774025 'location.href'
browser-agent session eval --session-id sess-xqbsmo --tab-id 216774025 \
  'JSON.stringify({title:document.title,url:location.href,text:(document.body&&document.body.innerText||"").slice(0,500)})'

# Run a local script file in the page
browser-agent session run --session-id sess-xqbsmo ./script.js
browser-agent session run --session-id sess-xqbsmo --tab-id 216774025 ./script.js

# Console / page logs
browser-agent session logs --session-id sess-xqbsmo
browser-agent session logs --session-id sess-xqbsmo --tab-id 216774025

# Screenshot
browser-agent session screenshot --session-id sess-xqbsmo -o /tmp/page.png
browser-agent session screenshot --session-id sess-xqbsmo --tab-id 216774025 -o /tmp/tab.png

# CDP (page-scoped; prefer create-tab over Target.createTarget)
browser-agent session cdp --session-id sess-xqbsmo Page.navigate '{"url":"https://example.com"}'
browser-agent session cdp --session-id sess-xqbsmo --tab-id 216774025 Page.navigate '{"url":"https://example.com"}'
browser-agent session cdp --session-id sess-xqbsmo --tab-id 216774025 Runtime.evaluate '{"expression":"document.title","returnByValue":true}'
browser-agent session cdp --session-id sess-xqbsmo --tab-id 216774025 Page.captureScreenshot '{}'
```

Typical task order:

```bash
browser-agent session new                                    # → sess-xqbsmo
browser-agent session info --session-id sess-xqbsmo
browser-agent session create-tab --session-id sess-xqbsmo https://example.com
# note tab_id from create-tab / info
browser-agent session eval --session-id sess-xqbsmo --tab-id 216774025 'document.title'
browser-agent session screenshot --session-id sess-xqbsmo --tab-id 216774025 -o /tmp/out.png
```

### 3. Tabs

- Run **`session info`** first: tab list, active, `job_target`, `recommended_cli`.
- Prefer **`--tab-id`** (stable). Default = active capturable tab in the session window.
- Avoid **`--tab-index`** (unstable). `create-tab` returns **`tab_id`** for later jobs.

### 4. `session cdp`

Page-scoped methods (`Runtime.*`, `Page.*`, DOM/CSS/Network as needed) OK.  
**Target.*** is a limited polyfill — use **create-tab** / **info**, not a full multi-target CDP graph.

