# browser-agent

Drive a live **Chrome** or **Firefox** browser from the terminal (or an AI agent):
open pages, run JavaScript, take screenshots, and more. A small control server
talks to a browser extension while you keep one **session page** open.

# Installation

Installs the **latest** GitHub Release for your OS/arch (no version flag required):

```sh
curl -fsSL https://raw.githubusercontent.com/xhd2015/browser-agent/master/install.sh | bash
```

This puts `browser-agent` on `$GOPATH/bin` (if that directory exists) or
`/usr/local/bin`. Ensure that directory is on your `PATH`.

Optional — pin an older release:

```sh
curl -fsSL https://raw.githubusercontent.com/xhd2015/browser-agent/master/install.sh \
  | INSTALL_TAG=v0.3.1 bash
```

# Quick start

Start a session (starts the control server on port **43761** if needed and opens
the browser):

```sh
browser-agent session new                 # Chrome (default)
browser-agent session new --browser firefox
```

Note the printed **`session-id`** (e.g. `sess-xxxxxx`). Reuse it on every later
command. Keep the session page open (`/go?session=…`) so the extension can stay
connected.

## Chrome

1. Install the extension package:

   ```sh
   browser-agent install-chrome-extension
   ```

2. Open `chrome://extensions` → enable **Developer mode** → **Load unpacked** →
   select the folder printed by the command.

3. Keep the session page open so the extension can connect.

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
```

## Agent skill

Copy the embedded skill into your agent tools (Claude Code, Codex, Grok, etc.):

```sh
browser-agent skill --show
browser-agent skill --install --global   # if supported by your skill host
```

# Tips

- Default control port: **43761**
- Prefer **`--tab-id`** from `session info` / `create-tab` for background tabs
- Do not navigate the session control page away from `/go?session=…` (that
  disconnects the extension)
- Long-running daemon only (optional): `browser-agent serve`

# License

See [LICENSE](LICENSE).
