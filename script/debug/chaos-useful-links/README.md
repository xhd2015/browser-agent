# chaos-useful-links

Real-Chrome **fuzzy chaos harness** for **browser-agent**. Seeds come from a
markdown/text link file (`--links`) or a built-in public catalog (`--random-links`).
No hardcoded `corpus.json`.

**Goal:** surface agent reliability issues (tab routing, attach, timeout, disconnect,
crash) under messy ops pages — not validate business correctness of the targets.

## Preconditions

1. `browser-agent` on `PATH` (or pass `--browser-agent /path/to/bin`)
2. Chrome extension **Load unpacked** once for your **default** Chrome profile
3. VPN / SSO as needed for internal hosts (when using private link files)
4. Keep the session control tab (`/go?session=…`) open during the run
5. Exactly one of `--links PATH` or `--random-links` is required

## Seed sources

| Flag | Meaning |
|------|---------|
| `--links PATH` | Extract http(s) URLs from a markdown/plain-text file (bare, `[t](url)`, backticks, `<>`; optional GFM table Env/Kind/Title/Market/Link) |
| `--random-links` | Built-in public https catalog (≥3 seeds: example.com, google.com, baidu.com) |

Mutex: cannot set both; neither is an error.

Optional extract options:

| Flag | Default | Meaning |
|------|---------|---------|
| `--include-archived` | false | Keep links under Historical/Archived/Deprecated headings |
| `--max-seeds N` | 0 | Cap seeds after dedupe (`0` = all) |
| `--kind` / `--env` | (none) | Filter when GFM table metadata is present |

Resolved snapshot is written to `out/<run-id>/corpus.resolved.json`.

Example fixtures (for dry-run / local parse checks):

- `tests/chaos-useful-links/testdata/mixed.md`
- `tests/chaos-useful-links/testdata/useful-links-table.md`
- `tests/chaos-useful-links/testdata/historical.md`

## Run

From this directory:

```sh
# Unit tests (dice + seedload)
go test .

# Dice plan only (no Chrome) — random public seeds
go run . --random-links --dry-run --seed 42 --max-ops 20

# Dice plan from a link file
go run . --links ../../../tests/chaos-useful-links/testdata/mixed.md --dry-run --seed 1 --max-ops 3

# Full chaos against default Chrome (opens session new unless --session-id set)
go run . --random-links --seed 42 --max-ops 40

# Reuse an existing session
export BROWSER_AGENT_SESSION_ID=sess-xxxx
go run . --links ../../../tests/chaos-useful-links/testdata/mixed.md \
  --session-id "$BROWSER_AGENT_SESSION_ID" --seed 7 --max-ops 30
```

From module root:

```sh
go run ./script/debug/chaos-useful-links --random-links --dry-run --seed 1 --max-ops 3
go run ./script/debug/chaos-useful-links \
  --links tests/chaos-useful-links/testdata/mixed.md \
  --dry-run --seed 1 --max-ops 3
```

### Flags

| Flag | Default | Meaning |
|------|---------|---------|
| `--links` | (none) | Path to markdown/text link file (mutex with `--random-links`) |
| `--random-links` | false | Use built-in public seed catalog (mutex with `--links`) |
| `--include-archived` | false | Include historical/archived/deprecated section links |
| `--max-seeds` | 0 | Cap after dedupe (`0` = all) |
| `--kind` / `--env` | (none) | Optional metadata filters |
| `--seed` | time-based | RNG seed (reproducible dice) |
| `--max-ops` | 40 | number of chaos steps |
| `--max-tabs` | 5 | max content tabs in session window |
| `--op-timeout` | 25s | per CLI job timeout |
| `--wait-extension` | 90s | wait after session new for extension connect |
| `--dry-run` | false | plan + fake tabs only |
| `--session-id` | (new) | reuse session |
| `--out` | `./out/<run-id>` | artifacts |
| `--json` | false | print `run.json` path only |

## Dice ops

Weighted random (constrained by tab count):

- `open_seed` — `session create-tab <url>`
- `navigate_seed` — `session cdp Page.navigate` on an existing content tab
- `eval_identity` — eval `location.href` / title / readyState with `--tab-id`
- `screenshot` — PNG under `out/.../screenshots/`
- `session_info` — dump `--json` to `session-info/`
- `logs` — console logs job
- `background_eval` — eval a non-active content tab via `--tab-id`
- `race_pair` — concurrent eval + screenshot

## Result classes

| Class | Meaning |
|-------|---------|
| `OK_LOADED` | identity eval looks like the content page |
| `OK_AUTH_WALL` | landed on SSO/login; agent still targeted the tab |
| `OK_SLOW` | succeeded but slow |
| `OK_INFO` | info/screenshot/logs ok |
| `FAIL_ROUTING` | eval hit `/go?session=` instead of content tab (**P1**) |
| `FAIL_TIMEOUT` | job hang (**P1**) |
| `FAIL_DISCONNECT` | extension dropped without navigating session page (**P1**) |
| `FAIL_ATTACH` | debugger attach failed (**P2**, sometimes expected) |
| `FAIL_CRASH` | daemon/connection death (**P0**) |

Exit **1** if any **P0/P1** issue; **0** otherwise (P2 still written to `issues/`).

## Artifacts

```text
out/<run-id>/
  corpus.resolved.json   # seeds + source meta + counts
  run.json
  issues/iss-001.json
  screenshots/step-01.png
  session-info/step-03.json
```

## Safety

- Read-only navigation / eval / screenshot only — no clicks on Deploy or DB write UIs
- Tabs opened only via `session create-tab` in the session window
- Does not navigate the `/go?session=` control tab away

## Replay

Same `--seed` + `--max-ops` + same seed source (`--links PATH` or `--random-links`, with the same extract flags) → same op sequence (Chrome timing may still vary classifications).
