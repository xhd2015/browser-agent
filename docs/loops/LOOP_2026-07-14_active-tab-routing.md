---
title: eval runs on session page tab while user tab is active
created: 2026-07-14
slug: active-tab-routing
path: docs/loops/LOOP_2026-07-14_active-tab-routing.md
loop_kind: bug-repro
establishment_status: VERIFY PASS
---

# LOOP: active tab routing mismatch

## Symptom (verbatim)

`session info` lists the user's tab as **active**, but `session eval` executes on the
**session control page** (`/go?session=...`) instead.

From operator session `sess-prd87k`:

```text
session info --json → browser.tabs[1] active:true
  title: Credit-消费贷 三流水线目标文件 - Google Sheets
  url: https://docs.google.com/spreadsheets/d/...

session eval '({ url: location.href, title: document.title })'
  → url: http://127.0.0.1:43761/go?session=sess-prd87k
  → title: Browser Agent Session
```

Docs (`SKILL.md`, `system_prompt.go`) claim eval/cdp target the **active tab in the
last-focused window**. Production extension `pickTargetTabIdForSession` pins to the
registered **session page tab** (`entry.tabId`).

## Root cause hypothesis

Phase 9 per-session routing over-corrected: jobs route to `/go?session=` tab for
multi-session safety, but never consult `entry.windowId` + active tab in that window.
Fixture extension (`browseragent/fixtures/extension/background.js`) still uses active-tab
routing — production diverged.

## Repro preconditions

1. `playwright-debug` on PATH (Chromium installed).
2. Worktree at `browser-agent-master-2026-07-14-refactor-browser-agent-to-client-control-server`.
3. Headed Chromium can launch (MV3 extension does not load in classic headless).

Do **not** apply code fixes in steps 1–4.

## Derived operations

| Op | Command |
|----|---------|
| D1 Build | `go build -o script/debug/active-tab-routing/browser-agent ./cmd/browser-agent` |
| D2 Trigger | `go run ./script/debug/active-tab-routing/main.go trigger` |
| D3 Inspect | `go run ./script/debug/active-tab-routing/main.go inspect` |
| D4 Playwright | `playwright-debug --extension <extDir> --headed run script/debug/active-tab-routing/testdata/active-tab-routing.js <baseURL> sess-loop-active-tab` |

## Pitfalls & blockers

| Pitfall | Mitigation |
|---------|------------|
| `playwright-debug` missing | BLOCKER — `go install` or smc plugin |
| Headless extension load fails | Use `--headed` (documented in e2e harness) |
| Network to example.com blocked | BLOCKER for trigger step |
| Navigating session tab away from `/go` | Disconnects WS — do not use as workaround |

## Aux script

`script/debug/active-tab-routing/main.go`

```sh
go run ./script/debug/active-tab-routing/main.go trigger   # writes evidence
go run ./script/debug/active-tab-routing/main.go inspect   # REPRO exit 1 when unfixed
```

Evidence: `script/debug/active-tab-routing/out/observation/playwright.json`

---

## 1. Build

```sh
cd /Users/xhd2015/.wrk/worktrees/browser-agent-master-2026-07-14-refactor-browser-agent-to-client-control-server
go build -o script/debug/active-tab-routing/browser-agent ./cmd/browser-agent
```

**Verify:** `test -x script/debug/active-tab-routing/browser-agent` → exit 0.

## 2. Deploy / Update

```sh
export BROWSER_AGENT_BIN="$PWD/script/debug/active-tab-routing/browser-agent"
export PATH="$HOME/go/bin:$PATH"
```

**Verify:** `command -v playwright-debug` → non-empty path.

## 3. Run (trigger failure only)

Starts ephemeral daemon, loads extension, opens session page + example.com tab
(active), posts eval job:

```sh
go run ./script/debug/active-tab-routing/main.go trigger
```

**Verify:** writes `script/debug/active-tab-routing/out/observation/playwright.json`;
playwright exits **non-zero** on unfixed tree.

## 4. Inspect / Feedback (assert symptom)

```sh
go run ./script/debug/active-tab-routing/main.go inspect
```

**Verify (bug-repro):** exit **1**; stdout contains `REPRO:` and shows eval URL
contains `/go?session=` while active tab URL contains `LOOP_MARKER=active-tab-routing`.

**Verify (post-fix / step 4b):** exit **0**; stdout contains `VERIFY:` and eval URL
contains `LOOP_MARKER=active-tab-routing`.

## 5. Fix (do not run during establishment)

In `Chrome-Ext-Browser-Agent/public/background.js` → `pickTargetTabIdForSession`:

1. Query `active: true` tab in `entry.windowId` (session window).
2. If capturable → use it for CDP jobs.
3. Fallback to registered session-page tab / `/go?session=` URL scan.

Sync `build/` + `browseragent/embedded/extension/background.js`. Update
`system_prompt.go` / `SKILL.md` to say **active tab in session window** (not global
last-focused). Remove harmful `Page.navigate` workaround on session tab.

**Post-fix verify:**

```sh
go run ./script/debug/active-tab-routing/main.go inspect
# expect exit 0, VERIFY:
```

---

## Run log

| Timestamp | Step | Result | Evidence |
|-----------|------|--------|----------|
| 2026-07-14T16:39+08 | 1 Build | PASS | `go build` exit 0 |
| 2026-07-14T16:39+08 | 2 Deploy | PASS | `playwright-debug` on PATH |
| 2026-07-14T16:40+08 | 3 Run | PASS | trigger wrote `playwright.json`; playwright exit 1 |
| 2026-07-14T16:40+08 | 4 Inspect | **SYMPTOM CONFIRMED** | inspect exit 1; `REPRO:` eval on `/go?session=` while active tab is `example.com/?LOOP_MARKER=active-tab-routing` |
| 2026-07-14T16:42+08 | 5 Fix + 4b | **VERIFY PASS** | inspect exit 0; `PASS:` eval on `example.com/?LOOP_MARKER=active-tab-routing` |

Establishment evidence (`playwright.json`):

```json
{"assert":"active_tab_routing","ok":false,"user_tab_active":true,
 "active_tab_url":"https://example.com/?LOOP_MARKER=active-tab-routing",
 "eval_url":"http://127.0.0.1:50113/go?session=sess-loop-active-tab",
 "bug_present":true}
```

---

## Handoff

Loop kind: **bug-repro**. Establishment: **VERIFY PASS** (fix landed 2026-07-14).

GREEN gate: `go run ./script/debug/active-tab-routing/main.go inspect` → exit 0, `PASS:`.