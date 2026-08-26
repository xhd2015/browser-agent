# browser-agent

Drive a live **Chrome** or **Firefox** browser from the terminal (or an AI agent):
open pages, run JavaScript, take screenshots, and more. A small control server
talks to a browser extension while you keep one **session page** open.

# Installation

Installs the **latest** GitHub Release for your OS/arch (no version flag required):

```sh
curl -fsSL https://raw.githubusercontent.com/xhd2015/browser-agent/master/install.sh | bash

# or install via go
go install github.com/xhd2015/browser-agent@latest
```

Then install the skills:

```sh
# Existing browser-driving skill (the default when no name is given)
browser-agent skill --install --global

# Browser-network-to-API workflow
browser-agent skill --install browser-agent-to-api --global
```

Then install browser extensions (details see below):
```sh
browser-agent install-chrome-extension
browser-agent install-firefox-extension
```

# Quick start

After you've done all the installations, then in agents (after skill install, in `claude code`, `codex`, `grok`, `opencode` etc.):

```md
use browser-agent to check what's new on my github trends
```

Agent will auto create a new session, open the browser and operate on behalf of you.

Or if you manually created a session:

```sh
# Chrome (default)
browser-agent session new

# or firefbox
browser-agent session new --browser firefox
```

Then  you can tell agent the session id, the agent will operate on that session:
```md
use browser-agent with session-id: sess-xxx to check what's new on my github trends
```

## Chrome

1. Install the extension package (on a TTY, macOS also auto Load unpacked via UI):

   ```sh
   browser-agent install-chrome-extension
   # extract + print only (CI / non-interactive):
   browser-agent install-chrome-extension --no-open
   # extract + write open handoff JSON (no Chrome UI):
   browser-agent install-chrome-extension --write-json-result /tmp/ba-ext.json
   ```

2. If UI automation did not run: open `chrome://extensions` → enable **Developer mode** → **Load unpacked** →
   select the folder printed by the command.

3. Keep the session page open so the extension can connect.

### Toolbar popup (optional)

The extension toolbar popup is only for status. CLI jobs do not require it.

**Note:** After Chrome has just started, the **first** open of the Browser Agent
popup can sometimes be slow (MV3 service worker cold start). This is intermittent;
later opens are usually quick. Prefer `browser-agent session info` for connection
status if the popup is slow.

## Firefox

1. Install the extension (on a terminal this also opens the signed `.xpi` in
   Firefox for a permanent install prompt):

   ```sh
   browser-agent install-firefox-extension
   # paths only (no launch):  browser-agent install-firefox-extension --no-open
   ```

2. **Permanent install (recommended)**
   - Confirm the Firefox install prompt from the opened `.xpi`, **or**
   - On the session page, click **Download / open browser-agent.xpi**
     (`http://127.0.0.1:43761/v1/firefox-xpi`), **or**
   - Use the shown `file:///…/browser-agent.xpi` URL (paste into the address bar
     if the browser blocks `file://` links from `http://` pages).

3. **Temporary add-on (fallback)** — unloads when Firefox restarts:
   - Open `about:debugging#/runtime/this-firefox`
   - **Load Temporary Add-on…** → open the printed folder under
     `…/browser-agent-firefox/<version>/` → select **manifest.json**

4. Keep the session page open so the extension can attach.

## Session commands

Reuse the same `--session-id` from `session new`:

```sh
browser-agent session info --session-id sess-xxxxxx
browser-agent session create-tab --session-id sess-xxxxxx https://example.com
browser-agent session eval --session-id sess-xxxxxx --tab-id <id> 'document.title'
browser-agent session screenshot --session-id sess-xxxxxx --tab-id <id> -o out.png

# Chrome only: capture all user HTTP(S) tabs and export per-tab HAR files
browser-agent session har start sess-xxxxxx
# perform the browser flow
browser-agent session har end sess-xxxxxx
# output: /tmp/browser-agent-sess-xxxxxx/manifest.json and tab-*.har

# Offline inspect (no live session)
browser-agent har inspect summary /tmp/browser-agent-sess-xxxxxx
browser-agent har inspect paths /tmp/browser-agent-sess-xxxxxx --host app.example.com
browser-agent har inspect show /tmp/browser-agent-sess-xxxxxx --match /api/example --json
```

## Agent skills

Discover, inspect, or install either embedded skill:

```sh
browser-agent skill --list

# Existing browser-driving skill; omitted name defaults to this skill
browser-agent skill --show
browser-agent skill --install --global

# Browser-to-API capture and analysis workflow
browser-agent skill --show browser-agent-to-api
browser-agent skill --install browser-agent-to-api --global
```

# Tips

- Default control port: **43761**
- Prefer **`--tab-id`** from `session info` / `create-tab` for background tabs
- Do not navigate the session control page away from `/go?session=…` (that
  disconnects the extension)
- Long-running daemon only (optional): `browser-agent serve`

### Chrome debugger notice (normal)

Chrome jobs use `chrome.debugger` on **one tab at a time** (the job target). While
that attach is live, Chrome often shows a security notice such as:

> Browser Agent started debugging this **browser**

That banner is **browser-wide UI** (it can appear on other windows, a Dock “New
Window”, Google, or even a blank New Tab). It does **not** mean every tab is a
CDP target, and it does **not** mean a window without a session page is being
driven. CDP still runs only on the attached tab; the session page
(`/go?session=…`) must stay open in the same window as the target for attach to
be allowed. The notice goes away after detach (session page closed, Cancel on
the bar, or sticky attach released).

# Development

To land a feature worktree, bump the version, tag, sign Firefox, and publish a
GitHub release, follow **[Publish / release](docs/publish.md)**.

# License

See [LICENSE](LICENSE).
