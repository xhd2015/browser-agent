---
title: session info 404 while serve --status lists session
created: 2026-07-14
slug: session-info-addr-mismatch
path: docs/loops/LOOP_2026-07-14_session-info-addr-mismatch.md
loop_kind: bug-repro
establishment_status: SYMPTOM CONFIRMED
---

# LOOP: session info addr mismatch

## Symptom (verbatim)

`serve --status` lists a live session, but `session info` for that same ID fails:

```text
$ browser-agent serve --status
...
Base URL: http://127.0.0.1:49976
...
Sessions (1)
Session ID    Phase
sess-rthgpj    waiting_extension

$ browser-agent session info --session-id sess-rthgpj
info failed: status 404: {"error":"session not found","message":"unknown session id"}
```

## Root cause hypothesis

- `serve --status` discovers the daemon via `~/.tmp/browser-agent/server.json` (actual addr, e.g. `:49976`).
- `session info` defaults to `http://127.0.0.1:43761` when `--addr` is omitted.
- A second daemon may also listen on `:43761` without the session → 404.

**Workaround (confirms hypothesis):** `session info --addr 127.0.0.1:49976` succeeds.

## Repro preconditions

1. `browser-agent serve` daemon **running** with `server.json` present under `~/.tmp/browser-agent`.
2. Daemon addr in `server.json` **differs** from default `127.0.0.1:43761` (ephemeral port is typical).
3. At least one session listed in `serve --status`.

Do **not** satisfy missing preconditions in steps 1–4; only in step 5 Fix.

## Observation notes

| Evidence | Value |
|----------|-------|
| `server.json` addr | `127.0.0.1:49976` (pid 73950) |
| Default port health | `curl :43761/v1/health` → `{"ok":true}` (pid 78979, stale second daemon) |
| `:43761/v1/sessions` | `404 page not found` (older API surface) |
| `:49976/v1/sessions` | lists `sess-rthgpj` |
| `session info` without `--addr` | hits `:43761` → 404 |
| `session info --addr 127.0.0.1:49976` | exit 0, returns session JSON |

## Pitfalls & blockers

| Pitfall | Mitigation |
|---------|------------|
| No daemon running | BLOCKER — run `browser-agent serve` or `session new` first |
| Daemon on default `:43761` only | Symptom may not reproduce; need ephemeral port or explicit `--addr` on serve |
| Wrong binary on PATH | Set `BROWSER_AGENT_BIN` to worktree build |
| Interactive prompts | None expected |

## Aux script

`script/debug/session-info-addr-mismatch/main.go`

```sh
go run ./script/debug/session-info-addr-mismatch/main.go
# REPRO: exit 1 when bug present
# exit 0 when session info works without --addr (post-fix verify)
```

---

## 1. Build

```sh
cd /Users/xhd2015/.wrk/worktrees/browser-agent-master-2026-07-14-refactor-browser-agent-to-client-control-server
go build -o script/debug/session-info-addr-mismatch/browser-agent ./cmd/browser-agent
```

**Verify:** `test -x script/debug/session-info-addr-mismatch/browser-agent` → exit 0.

## 2. Deploy / Update

Use the worktree-built binary (not necessarily `$PATH`):

```sh
export BROWSER_AGENT_BIN="$PWD/script/debug/session-info-addr-mismatch/browser-agent"
```

**Verify:** `"$BROWSER_AGENT_BIN" serve --status | head -1` → contains `browser-agent daemon status`.

## 3. Run (trigger failure only)

No `--addr` on session info — this is the operator path that fails:

```sh
SESSION_ID=$( "$BROWSER_AGENT_BIN" serve --status | awk '/^sess-/ {print $1; exit}' )
"$BROWSER_AGENT_BIN" session info --session-id "$SESSION_ID"
```

**Verify:** command exits **non-zero**; stderr/stdout contains `session not found` or `status 404`.

## 4. Inspect / Feedback (assert symptom)

```sh
export BROWSER_AGENT_BIN="$PWD/script/debug/session-info-addr-mismatch/browser-agent"
go run ./script/debug/session-info-addr-mismatch/main.go
```

**Verify (bug-repro):** exit **1**; stdout contains `REPRO:` and `session not found` or `status 404`.

Optional evidence capture:

```sh
mkdir -p script/debug/session-info-addr-mismatch/out/observation
"$BROWSER_AGENT_BIN" serve --status > script/debug/session-info-addr-mismatch/out/observation/status.txt
"$BROWSER_AGENT_BIN" session info --session-id "$SESSION_ID" \
  > script/debug/session-info-addr-mismatch/out/observation/info-default.txt 2>&1 || true
```

## 5. Fix (do not run during establishment)

Resolve control-plane addr for session side-commands from `server.json` (same as `serve --status` / `EnsureDaemon`):

- When `--addr` is omitted on `session info|eval|run|logs|screenshot|cdp`, read `{base-dir}/server.json` and use recorded `base_url` / `addr`.
- Default `--base-dir` to `~/.tmp/browser-agent` (match status).
- Consider warning when multiple daemons listen (stale orphan on `:43761`).

**Post-fix verify (for `/loop-workflow`):**

```sh
go run ./script/debug/session-info-addr-mismatch/main.go
# expect exit 0, stdout contains VERIFY:
```

---

## Run log

| Timestamp | Step | Result | Evidence |
|-----------|------|--------|----------|
| 2026-07-14T13:50+08 | 1 Build | PASS | `go build` exit 0 |
| 2026-07-14T13:50+08 | 2 Deploy | PASS | `BROWSER_AGENT_BIN` set to worktree binary |
| 2026-07-14T13:50+08 | 3 Run | PASS | `session info` exit 1, `session not found` |
| 2026-07-14T13:50+08 | 4 Inspect | **SYMPTOM CONFIRMED** | inspect script exit 1, `REPRO: ... sess-rthgpj at http://127.0.0.1:49976` |

---

## Handoff

Loop kind: **bug-repro**. Establishment: **SYMPTOM CONFIRMED**.

Next: `/loop-workflow` — auto-resolve `--addr` for session side-commands from `server.json`; GREEN gate = inspect script exit 0 with `VERIFY:`.