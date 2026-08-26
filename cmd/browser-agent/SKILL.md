---
name: browser-agent
description: >-
  Use when research a browser or web page, or a non-api url mentioned. Drive a live Chrome session: start with
  session new (auto-ensures daemon), open the session page, then use nested
  session side commands (session info, create-tab, eval, run, logs, screenshot,
  cdp, har).
metadata:
  version: "1.0.15"
---

# Browser Agent Skill

Drive **browser-agent** (control **127.0.0.1:43761**): **`session new` once** → keep session page → side commands with the **same** `--session-id`.

Do not run `open-managed-chrome` unless the user explicitly asks for a separate managed Chrome profile. It creates a new Chrome profile/window and is not part of the normal session workflow.


## Flow

### 1. `session new` (once per task)

```bash
browser-agent session new
```

Chrome is the default browser. For **Firefox**, pass `--browser firefox`:

```bash
browser-agent session new --browser firefox
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

If `session new` waits for the extension and times out, stderr ends with:

```text
Please run or ask user to run manually: this needs user handling
```

plus `install-chrome-extension` (or `install-firefox-extension`) and manual load steps. **That needs user handling:** run the install command or ask the user to Load unpacked / temporary add-on. Then reuse the **same** session-id (`session info` / jobs) — do **not** call `session new` again.

### 1b. Extension install (Chrome vs Firefox)

**Chrome** (default):

```bash
browser-agent install-chrome-extension
```

On a **TTY** (macOS), the CLI also drives **Load unpacked** via UI automation
(Accessibility), then **best-effort removes older same-name** extension cards.
Use **`--no-open`** to extract and print the path only (default for pipes/CI).
**`--write-json-result FILE`** writes open-handoff JSON and implies `--no-open`.
**`--keep-old`** skips older-card removal. Other flags: `--open`, `--dry-run`,
`--dump-tree`, `--extension-dir`.

If UI fails, open `chrome://extensions` → Developer mode → Load unpacked → path
printed by the CLI.

**Firefox**:

```bash
browser-agent install-firefox-extension
```

When a signed `.xpi` is embedded, the CLI prints the **xpi path** and (on a TTY)
opens it in Firefox so you can confirm the permanent install prompt.

On the **session page** (Firefox), also use:

- **Download / open** `http://127.0.0.1:43761/v1/firefox-xpi` (reliable click-to-install)
- The shown **`file:///…/browser-agent.xpi`** URL (paste into the address bar if
  `file://` navigation from the session page is blocked)

Fallback — temporary add-on (unloads on Firefox restart):

1. Open `about:debugging#/runtime/this-firefox` (or type `about:debugging` → This Firefox)
2. Click **Load Temporary Add-on…**
3. Open the folder under `…/browser-agent-firefox/<version>/`
4. Select **manifest.json** (not the folder), then Open

Use `--no-open` to print paths only (no Firefox launch).

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

# Chrome HAR capture (all user HTTP(S) tabs in the session window)
browser-agent session har start sess-xqbsmo
# perform the browser flow, then export per-tab HAR files plus manifest.json
browser-agent session har end sess-xqbsmo
browser-agent session har end sess-xqbsmo --output-dir /tmp/my-har-capture

# Offline HAR inspect (no live session)
browser-agent har inspect summary /tmp/browser-agent-sess-xqbsmo
browser-agent har inspect paths /tmp/browser-agent-sess-xqbsmo --host app.example.com
browser-agent har inspect show /tmp/browser-agent-sess-xqbsmo --match /api/example --json

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

### 5. Daemon probe

Read-only status (no spawn):

```bash
browser-agent serve --status
```

Prefer **`session new`** (auto-ensures daemon) over legacy **`serve --session-id`** bootstrap; plain **`serve`** remains for long-running daemon hosts.
