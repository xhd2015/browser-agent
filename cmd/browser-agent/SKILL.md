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

Drive **browser-agent** (control **127.0.0.1:43761**): **`session new` once** → keep the `/go?session=` tab → side commands with the **same** `--session-id`.

## Session lifecycle

```bash
browser-agent session new
# Chrome default. Firefox: --browser firefox
# Isolated Chrome (cookie import / HAR; does not touch default profile):
#   browser-agent session new --adhoc-browser-profile
```

Stdout includes `session-id: sess-…`. Reuse that id on every later command.

| Do | Don't |
|----|--------|
| `session create-tab` for new pages | Call `session new` again to “refresh” |
| Keep the `/go?session=<id>` tab open | Close it or navigate that tab away |
| Second `session new` only if the first is gone/unusable or isolation is required | Spam `session new` when attach stalls |

On extension attach timeout, stderr asks for user handling (`install-chrome-extension` / `install-firefox-extension` + Load unpacked / temporary add-on). **Reuse the same session-id** after install — do not `session new` again.

Profiles: default Chrome/Firefox for normal work; **`--adhoc-browser-profile`** for Firefox→Chrome cookie import (Chrome-only). Do **not** run `open-managed-chrome` unless the user asks (persistent managed profile, not the normal flow). Cookie playbook: `browser-agent how-to credentials/export`.

## Extension install

**Chrome:** `browser-agent install-chrome-extension` (TTY macOS can Load unpacked via UI). Flags: `--no-open`, `--write-json-result FILE`, `--keep-old`, `--open`, `--dry-run`, `--dump-tree`, `--extension-dir`. Fallback: `chrome://extensions` → Developer mode → Load unpacked → printed path.

**Firefox:** `browser-agent install-firefox-extension` (embedded `.xpi` when present). On the session page: `/v1/firefox-xpi` or the shown `file:///…/browser-agent.xpi`. Temporary add-on: `about:debugging` → This Firefox → Load Temporary Add-on → `manifest.json` under `…/browser-agent-firefox/<version>/`.

## Side commands

Assume `session-id: sess-xqbsmo` and a content `tab_id` from `session info` / `create-tab`.

```bash
browser-agent session info --session-id sess-xqbsmo
browser-agent session info --session-id sess-xqbsmo --json
browser-agent session create-tab --session-id sess-xqbsmo https://example.com
browser-agent session eval --session-id sess-xqbsmo --tab-id <tab> 'document.title'
browser-agent session run --session-id sess-xqbsmo --tab-id <tab> ./script.js
browser-agent session logs --session-id sess-xqbsmo --tab-id <tab>
browser-agent session screenshot --session-id sess-xqbsmo --tab-id <tab> -o /tmp/out.png
browser-agent session har start sess-xqbsmo
browser-agent session har end sess-xqbsmo
browser-agent har inspect summary /tmp/browser-agent-sess-xqbsmo
browser-agent session cdp --session-id sess-xqbsmo --tab-id <tab> Page.navigate '{"url":"https://example.com"}'
```

## Tabs

- Run **`session info` first** (tab list, `job_target`, `recommended_cli`).
- Prefer **`--tab-id`** (stable). Avoid **`--tab-index`**.
- `create-tab` returns **`tab_id`** for later jobs.

## CDP

Page-scoped methods (`Runtime.*`, `Page.*`, DOM/CSS/Network) OK. **`Target.*`** is a limited polyfill — use **create-tab** / **info**, not a full multi-target CDP graph.

## Daemon

```bash
browser-agent serve --status   # read-only; no spawn
```

Prefer **`session new`** (auto-ensures daemon) over legacy **`serve --session-id`**. Plain **`serve`** remains for long-running daemon hosts.
