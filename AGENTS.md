# Agent notes — browser-agent

## Installing a local binary (fat embed)

Prefer:

```bash
go run ./script/browser-agent/install
# or: go run ./script/browser-agent/install --skip-firefox-sign
```

Do **not** use bare `go install ./cmd/browser-agent` or `go build` when you need
the Chrome extension / React session-page `//go:embed` tree.

After install, restart any running daemon:

```bash
browser-agent serve --stop
browser-agent session new
```

Then reload the Chrome extension if its files changed
(`chrome://extensions` → Browser Agent → Reload).

## Attach / connection debugging

Each session appends JSONL events to `sessions/<id>/log.jsonl` (create, chrome
open, attach stages, WS hello/disconnect, wait timeout, SW wakeup). Inspect offline:

```bash
browser-agent session log --session-id sess-xxxxxx
browser-agent session log --session-id sess-xxxxxx --json --tail 50
```

By default `session new` opens a **quiet background wakeup tab** on attach stall
(look for `sw_wakeup_open` in the log). Disable with
`browser-agent session new --no-wakeup-workaround` (then use the toolbar popup /
Reload if the SW stays asleep).

While connected, the extension keeps the MV3 worker alive with an **offscreen**
document + `/go` Port pings, scopes heal to control tabs, and times out debugger
attach so eval/screenshot cannot hang the worker forever. After install/Reload,
re-grant permissions if Chrome prompts for `offscreen`.

Extension SW `baLog` lines are POSTed to `/v1/ext/log` and appended to a
**unified** file (not per-session):

```bash
browser-agent extension log --browser chrome
# → ~/.tmp/browser-agent/extension-chrome.log.jsonl
browser-agent extension log --browser firefox --tail 50
```

(`session logs` is still the live console-buffer job and needs a connected extension.)