---
name: browser-trace
description: >-
  Capture browser network traffic with the browser-trace CLI, wait for the user
  to finish their flow, then inspect the session directory (recording.har +
  meta.json) and discuss findings against the user's request. Use when user explicitly mentioned
---

# Browser Trace Skill

Agent playbook for **run → wait → inspect session → analyse → discuss**.

`browser-trace` starts a local control server, opens Chrome (new window), records
network traffic from all tabs in that window via **Chrome-Ext-Capture-API**, and
on stop writes artifacts under a session directory. **Stdout is only the session
path** (one line). Progress goes to stderr / `browser-trace.log`.

## When to use

- User wants to capture real browser network for a task they will perform
- You need live HAR evidence before changing API clients or diagnosing a UI flow
- User mentions browser-trace, HAR capture, or Chrome-Ext-Capture-API recording

Do **not** invent endpoints when a short capture would answer the question.

## Prerequisites

1. **`browser-trace` on PATH** (preferred)
2. **Extension available** — CLI embeds the extension and best-effort

Listen address defaults to `127.0.0.1:43759` and **fails if busy**.

## Agent workflow (required)

Follow these steps in order whenever the user asks you to do something that
needs a browser-trace capture.

### 1. Clarify the request (briefly)

Confirm in one pass if missing:

- **Goal** — what to learn or fix (e.g. “which API updates subtask status”)
- **User actions** — what they will do in the browser (ordered steps)
- **Hosts of interest** (optional) — e.g. `app.example.com`

If the goal is clear enough to capture, do not block on perfect detail.

### 2. Start browser-trace and wait for the user

Run the CLI in the **foreground** and **wait until it exits**. Do not kill it
early; the user completes their flow and stops recording.

```bash
browser-trace
# optional:
# browser-trace --verbose
# browser-trace --base-dir /tmp/browser-trace-sessions
```

Tell the user clearly:

1. A Chrome window should open (or they use the session URL if already open).
2. Perform the agreed actions in **that** window’s tabs (https pages capture).
3. Stop when done: **Stop** on the extension popup, or **Ctrl-C** in the terminal.
4. You will read the session directory printed when the process exits.

**While waiting:** keep the process running (long timeout / interactive). Parse
**stdout** when it exits — a single absolute path, e.g.:

```text
/Users/you/.tmp/browser-trace/2026-07-11-14-30-00-abc
```

Stderr may show ready / saved / heartbeat warnings; treat a non-zero exit as
failure and surface the error before analysing.

### 3. Check collected files under the printed directory

Let `SESSION` be the path from stdout. Verify:

| Path | Role |
|------|------|
| `$SESSION/recording.har` | HAR 1.2 network log (primary evidence) |
| `$SESSION/meta.json` | Session status, `entry_count`, `stop_reason`, optional `partial` |
| `$SESSION/browser-trace.log` | Lifecycle log (unless `--no-log-file` / quiet) |

```bash
ls -la "$SESSION"
python3 -m json.tool "$SESSION/meta.json"
```

**meta.json checklist:**

- `status` — expect saved success path; note `failed` / errors
- `stop_reason` — `extension`, `signal`, `heartbeat_lost`, …
- `entry_count` — 0 means empty capture (wrong window, no attach, or no traffic)
- `partial: true` — incomplete snapshot (e.g. heartbeat lost); still usable, flag confidence
- `extension_version` / `supports_browser_trace` — extension capability

If `recording.har` is missing or `entry_count` is 0, diagnose before deep analysis
(extension not recording, chrome:// only tabs, wrong window, port conflict).

### 4. Analyse the trace

Goal: map **user narrative → relevant requests → contract / gaps**, then answer
the original request.

1. **Summarize entries** — method, host, path, status, timing; prefer tools over
   dumping the full HAR.
2. **Filter noise** — drop static assets, analytics, preflight unless relevant.
3. **Keep** app API calls that match the user’s steps (POST/PUT/PATCH/DELETE first;
   GET when part of the flow).
4. **Align timeline** — order by `startedDateTime`; group into transactions.
5. **Extract contracts** for key calls: method+path, request body, response
   envelope, auth style (do not paste secrets), success signal.
6. If the repo has **`analyse-har`** (or `skills/analyse-har`), use its
   `summarize_har.py` and deeper playbook when reverse-engineering APIs:

   ```bash
   python3 skills/analyse-har/scripts/summarize_har.py \
     "$SESSION/recording.har" --host <app-host> --json
   ```

### 5. Discuss with the user (analysis-first)

Present findings **before** implementing unless they already asked to implement.

Suggested structure:

```markdown
## Request
<one sentence restating the user goal>

## Capture
- Session: <SESSION path>
- Stop: <stop_reason> · entries: <n> · partial: <yes/no>
- Extension: <version if present>

## What the browser did (relevant)
1. `METHOD path` → status · short note
2. …

## Analysis vs your request
- <answer, gap table, or root-cause hypothesis>

## Options / next steps
- Re-record with … / implement fix / dig into entry N …
```

**Discuss:** ask if the sequence matches what they did, whether hosts look right,
and what they want next (deeper dive, code change, another capture).

Do **not** silently edit product code from this skill alone; propose first unless
the user already ordered an implementation.

## CLI reference (runtime, not skill install)

```bash
browser-trace [options]
browser-trace --install-chrome-extension
browser-trace --addr 127.0.0.1:43759
browser-trace --base-dir ~/.tmp/browser-trace
browser-trace --ready-timeout 30s
browser-trace --no-open-chrome    # tests / attach manually
browser-trace -v / --verbose
browser-trace --quiet
```

Stop: Ctrl-C or extension popup **Stop**. Success → session path on stdout;
artifacts: `recording.har` + `meta.json`.

## Failure / empty capture tips

| Symptom | Likely cause | What to try |
|---------|--------------|-------------|
| Listen fail / address in use | Port 43759 busy | Stop other browser-trace; or `--addr` |
| Ready timeout | Extension never hello/start | Install/reload extension; close window; re-run |
| `entry_count: 0` | No capturable traffic / not attached | Use https app tabs; reload ext; check popup TABS |
| `heartbeat_lost` + partial | Browser/extension gone without Stop | Use partial HAR; re-record for full complete |
| New tab missing from capture | Attach policy / chrome:// only | Navigate to http(s); ensure recording window |

## Example

User: “Capture what Casement does when I mark a subtask DONE, then tell me which APIs we need.”

Agent:

1. Confirm: weekly panel → status TO DO → DONE.
2. Run `browser-trace`; tell user to act in the opened window, then Stop.
3. Read stdout session path; open `meta.json` + `recording.har`.
4. Summarize/filter to the target app host; map multi-step API sequences.
5. Discuss contract + any gap vs repo clients; wait for implement go-ahead.
