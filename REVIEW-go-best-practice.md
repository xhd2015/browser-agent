# Go Best Practice Review — browser-agent

**Date:** 2026-08-06  
**Scope:** Codebase structure, CLI design, flag handling, package layout  
**Method:** Inspected tree against [go-best-practice](https://github.com/xhd2015) skill topics: `flags-parsing` (+ subcommand/types), `cli` (+ color/skill-cli/streaming/dry-run), `cmd-exec`, `go-embed-assets`, `kool-create`  
**Status:** Review only — no production fixes applied.

---

## Executive summary

The project is a multi-binary monorepo (`browser-agent`, `browser-trace`, scripts, extensions, React UIs) with a **strong** asset-embed/hydrate story and a **partial** migration to less-flags + cli-color. The highest-impact gaps are:

1. **Inconsistent flag parsing** — only `serve` and `session list` use `less-flags`; most operator commands still use hand-rolled `flagString` / `flagBool` that **silently ignore typos**.
2. **Help is not per-command** — most `session * --help` paths dump the global `fullHelp` megablob instead of a command-local surface (`flags-parsing/subcommand`).
3. **Monolithic `browseragent` package** (~14k LOC, 67 files) mixes CLI, daemon, HTTP control plane, embed hydrate, and build/bundle orchestration with a large public API surface.

Asset hydrate, skill Shape‑1 packaging, release `--dry-run`, and `serve`/`session list` less-flags paths are already aligned with the skill and are good templates for the rest of the CLI.

---

## What’s already strong

| Area | Evidence | Skill topic |
| ---- | -------- | ----------- |
| **go:embed + hydrate** | Placeholder-only git roots, gitignore un-ignore of `placeholder.txt`, `EmbedComplete*`, cache under XDG/`~/.cache`, version-pinned release tar.gz, `assets ensure\|status`, `docs/assets-hydrate.md` | `go-embed-assets` |
| **Skill CLI Shape 1** | `//go:embed SKILL.md`, `skillcmd.SingleSkill`, `browser-agent skill --show\|--list\|--install`; browser-trace mirrors | `cli/skill-cli` |
| **less-flags (partial)** | `parseServeOptions`, `parseSessionListOptions`: `HelpFunc` + `HelpNoExit`, unknown-flag errors with “Run … --help” | `flags-parsing` |
| **Color policy (partial)** | `cli_color.go`: `--color` / `--no-color` conflict, `NO_COLOR` in auto, `term.IsTerminal`; scripts use `script/internal/clicolor` | `cli/color` |
| **Release dry-run** | `script/github/release`: one pipeline, soft-fail tag/creds in dry-run, `[dry-run]` plan lines, same planned names as live | `cli/dry-run` |
| **cmd-exec (scripts)** | `install`, `browser-trace/bundle`, chrome-ext build, har-viewer build: `cmd.Debug().Dir(...).Run(...)` with inherited stdout/stderr | `cmd-exec` |
| **Thin mains** | `cmd/browser-agent` only wires signals for `serve` and delegates to `browseragent.HandleCLI` | package layout hygiene |

---

## Findings (severity-ordered)

### High

#### H1. Hand-rolled flag helpers on most session/job CLIs; unknown flags silently ignored

**Where:** `browseragent/cli.go` (`flagString`, `flagBool`, `flagStringSet`, `hasHelpFlag`, `takePositional`), used by `session new`, `info`, `delete`, `eval`, `run`, `logs`, `screenshot`, `cdp`, `create-tab`, `open-managed-chrome`, install-chrome (partial), assets help only.

**Contrast:** `serve` and `session list` already use `github.com/xhd2015/less-flags` and reject unrecognized flags.

**Why it matters (`flags-parsing`, `flags-parsing/types`):**

- Typos like `--sesion-id`, `--base_dir`, or `--timeuot` are **ignored**; jobs may use wrong defaults or env session with no error.
- Hand parsers reimplement equals-form (`--flag=value`), bool `=true`/`=false`, and value consumption for positionals — easy to drift from less-flags behavior.
- `takePositional` hard-codes a skip-list of flag names; new flags added without updating the list break positional arg extraction.
- Invalid `--timeout` values fall through to the default in `jobTimeoutMS` instead of failing parse (`flags-parsing/types` prefers typed `Duration` / explicit validation).

**Recommended change:**

1. Define small parse helpers per command (or shared option structs) with less-flags, e.g.:

   ```go
   remain, err := lessflags.
       String("--session-id", &sessionID).
       String("--base-dir", &baseDir).
       String("--host", &host).
       String("--server-port", &portStr).
       String("--tab-id", &tabID).
       String("--tab-index", &tabIndex).
       Duration("--timeout", &timeout). // or String + validate
       Help("-h,--help", sessionEvalHelp).
       HelpNoExit().
       Parse(args)
   ```

2. Prefer `**string` / pointer targets when “unset vs empty” matters (session-id env vs flag).
3. Delete `flagString` / `flagBool` once call sites migrate; keep only less-flags + remain positionals.
4. Mirror serve’s error suffix: `Run 'browser-agent session eval --help' for usage.`

**Migration order (suggested):** `session new` → job commands (`eval`/`run`/…) → `session info`/`delete` → `open-managed-chrome` / firefox install hand loop → drop helpers.

---

#### H2. Subcommand help dumps global `fullHelp` instead of level-specific help

**Where:** `cliSession` and nearly all `cliEval` / `cliRun` / `cliScreenshot` / `cliSessionNew` / `cliInfo` / `cliOpenManagedChrome` paths: `hasHelpFlag` → write `fullHelp`.

**Contrast:** `serve` has `serveHelp`; `session list` has `sessionListHelp`; `assets` has `assetsHelp`.

**Why it matters (`flags-parsing/subcommand`):**

- Users running `browser-agent session eval --help` need eval-specific usage (expr, tab-id, timeout defaults), not the entire product manpage.
- Recipe rule: **every command level** answers `-h`/`--help` with **that level’s** usage; top-level help should say *Run `browser-agent <command> --help` for command-specific options.*
- Root `fullHelp` already documents everything, which works as a dump but fights progressive discovery and will diverge from real flags as parsers evolve.

**Recommended change:**

1. Add constants: `sessionHelp` (dispatch), `sessionNewHelp`, `sessionEvalHelp`, … or generate from the same struct as less-flags.
2. Wire `lessflags.Help` / `HelpFunc` per handler (not a manual `hasHelpFlag` scan — recipe says parse with less-flags only for color/flags).
3. Shrink root help to command index + globals; point to subcommand help.
4. Empty `session` with no subcommand: print `sessionHelp` (friendly default) instead of only briefUsage + error.

---

### Medium

#### M1. Two flag libraries in one module: `less-flags` vs `less-gen/flags`

**Where:**

| Path | Library |
| ---- | ------- |
| `browseragent/serve.go`, `session_list.go`, `script/github/release` | `github.com/xhd2015/less-flags` |
| `cmd/browser-trace/main.go`, `har-viewer/run/run.go` | `github.com/xhd2015/less-gen/flags` |

**Why it matters (`flags-parsing`):** Different packages, help exit semantics (`ErrHelp`), and type APIs confuse contributors and force dual mental models. Module already depends on both.

**Recommended change:** Standardize on **`less-flags`** for all product CLIs (browser-agent + browser-trace + har-viewer). Drop `less-gen/flags` from product entrypoints when convenient. Keep hand-rolled only for one-off script arg peels that `script/internal/clicolor.ParseFlags` already does.

---

#### M2. Color implementation split; incomplete flag surface on human commands

**Where:**

- Product: `browseragent/cli_color.go` (`newServeColor`)
- Scripts: `script/internal/clicolor` (ColorMode, Resolve, FlagHelp, Style)
- `session info` human path: `newServeColor(stdout, env, false, false)` — **no** `--color` / `--no-color` even though help documents color for serve/list
- Firefox install: hand-parses color flags; serve/list use less-flags

**Why it matters (`cli/color`):**

- Skill prefers one ColorMode model + less-flags bools; conflict message is already correct on paths that check both flags.
- Duplicating resolve logic risks NO_COLOR / TTY drift (`noColorEnvSet` uses `TrimSpace`; skill’s `os.Getenv("NO_COLOR") != ""` treats whitespace-only as set — minor, but keep one implementation).
- Operators cannot force color on `session info` tables when piping through a pager that is still a TTY, or force off without env.

**Recommended change:**

1. Lift `script/internal/clicolor` (or a `browseragent`/`internal/clicolor` package) for **both** product and scripts, or move product color into a small shared package under `internal/`.
2. Add `--color` / `--no-color` to every human table command (`session info`, status-like output) via less-flags.
3. Keep machine `--json` colorless (already good).

---

#### M3. Monolithic `package browseragent` (~14k LOC) without internal boundaries

**Layout today:**

```text
cmd/browser-agent/main.go          → thin entry
browseragent/                      → CLI + daemon + HTTP + WS + jobs + embed +
                                     hydrate + bundle/build + firefox + skill
browseragent/inject/               → test hooks only
browsertrace/                      → second product package
script/…, react/, Chrome-Ext-*, Firefox-Ext-*, har-viewer/
```

**Why it matters (package layout / Go practice):**

- ~174 exported symbols; CLI tests and library consumers share one package, so accidental coupling is easy.
- Build/bundle (`bundle.go`, `build_assets.go`) lives next to the operator runtime — increases risk of pulling Node-oriented helpers into the shipped mental model.
- Harder to enforce “library API vs CLI-only” and to reuse smaller pieces from scripts without importing the world.

**Recommended change (incremental, not big-bang):**

| Package (suggested) | Responsibility |
| ------------------- | -------------- |
| `browseragent` or `internal/control` | Daemon, sessions, jobs, HTTP/WS |
| `internal/cli` or keep `HandleCLI` thin | Arg parse + dispatch only |
| `internal/assets` / keep assets_* | Embed complete + hydrate + cache |
| `internal/bundle` or stay under `script/` | Staging/build for embed (already partly script-driven) |

`cmd/*/main.go` stay thin. Prefer **`internal/`** for non-module-public types. Do not block features on a full split; migrate when touching adjacent code (H1 is a natural CLI extraction moment).

**kool-create note:** Existing `react/` + Go backend is conceptually close to `kool create go-react`, but the product evolved past a scaffold. No need to re-scaffold; use kool only for **new** side apps (e.g. har-viewer frontends).

---

#### M4. External commands: mixed `os/exec` vs `xgo/support/cmd`

**Where:**

| Style | Examples |
| ----- | -------- |
| `cmd.Debug().Dir().Run` | `script/browser-agent/install`, `script/browser-trace/*`, `script/chrome-ext-capture-api/build`, `har-viewer/script/*`, `script/github/release` |
| Raw `exec.Command` + manual Stdout/Stderr | `script/browser-agent/bundle` (`runGenerate`), `script/github/release-assets` (`gh`), firefox sign/bundle, bump-version git, product `browser_bin.go` / `agent_run.go` / `session_new.go` / `firefox.go` |

**Why it matters (`cmd-exec`):**

- Recipe prefers `cmd.Debug()` for developer scripts so debug mode, Dir, Env, and I/O inherit consistently.
- Product launches of Chrome/Firefox/agent-run legitimately use `os/exec` (Start/detach, no Debug noise) — **acceptable** if intentional.
- Script inconsistency is the real issue: e.g. browser-agent **install** uses `cmd.Debug`, **bundle** uses bare `exec.Command` for generate.

**Recommended change:**

1. Scripts that chain `go run` / `npm` / `gh`: migrate to `cmd.Debug().Dir(root).Run(...)` (or `.Output()` when capture is required).
2. Document product policy: *browser process launches stay on `os/exec`; build/release scripts use `support/cmd`.*
3. When capture is needed (`go env GOPATH`), prefer `cmd.Output` over ad-hoc `exec.Command(...).Output()` for consistency.

---

#### M5. Help / behavior doc drift in one blob

**Examples:**

- Root `fullHelp` says `serve --status` “prints JSON”; `serveHelp` says “human table”.
- Root help lists all flags but does not say *Run `browser-agent <command> --help`…* (`flags-parsing/subcommand` checklist).
- `install-chrome-extension` accepts `--base-dir` via `flagString` but does not reject unknown flags or offer command-local help.

**Recommended change:** Fix status wording when adding serve-specific tests; adopt per-command help (H2) so drift is confined; reject unknown flags on install commands.

---

### Low

#### L1. `har-viewer` embed of `dist` without placeholder layer

**Where:** `har-viewer/main.go` `//go:embed project-api-capture-react/dist`.

**Why it matters (`go-embed-assets`):** Clean clones without a prior frontend build can fail `go build` with “no matching files found”. browser-agent/browser-trace already follow the five-layer pattern; har-viewer is a secondary tool.

**Recommended change:** Placeholder + gitignore + optional hydrate, or document “must build frontend before `go build ./har-viewer`” in README (minimum). Prefer placeholders if har-viewer is ever published via `go install`.

---

#### L2. Firefox extension shell fully tracked in git

**Where:** `browseragent/embedded/extension-firefox/**` (full JS/manifest) vs Chrome/session-page/xpi placeholders only.

**Why it matters (`go-embed-assets`):** Intentional “P1 shell” per comments, but dual policies confuse operators and increase module zip size. Completeness checks may treat it differently from Chrome.

**Recommended change:** Document the exception in `docs/assets-hydrate.md` (already partially Chrome/session focused). Long-term: same placeholder + bundle stage if the Firefox shell is generated.

---

#### L3. Streaming / buffering on job CLIs

**Where:** `postJobAndPrint` reads full HTTP body then prints one JSON line.

**Why it matters (`cli/streaming`):** Fine for single job results (small fixed payload). No change required unless you add long-running multi-result streams — then prefer NDJSON or progressive stderr milestones (browser-trace already streams lifecycle to stderr).

---

#### L4. Dry-run coverage outside release

**Where:** `script/github/release` and `script/bump-version` / chaos debug tools implement dry-run; operator `browser-agent` has no dry-run (correct — no multi-step destructive plan CLI).

**Note:** `serve --stop` is interactive confirm, not dry-run — OK.

---

## Package / CLI map (current)

```text
browser-agent
├── serve                 [less-flags]  color, stop/status/kill-existing
├── session
│   ├── new               [hand flags]  fullHelp on --help
│   ├── list              [less-flags]  dedicated help
│   ├── info|delete|…     [hand flags]  fullHelp on --help
│   └── eval|run|… jobs   [hand flags]  silent unknown flags
├── install-*-extension   [mixed hand]
├── open-managed-chrome   [hand flags]
├── skill                 [skillcmd Shape 1]
└── assets ensure|status  [dispatch + assetsHelp]

browser-trace
├── (default capture)     [less-gen/flags]
├── skill                 [skillcmd]
└── assets                [browsertrace.HandleCLI]
```

---

## Recommended change backlog (grounded)

Prioritized for impact vs. skill alignment. **Do not implement in this review pass.**

| Priority | Change | Topics |
| -------- | ------ | ------ |
| P0 | Migrate session job + `session new` flags to less-flags; reject unknown flags | `flags-parsing`, `types` |
| P0 | Per-command help strings + help at every dispatch level | `flags-parsing/subcommand` |
| P1 | Unify browser-trace (and har-viewer) on less-flags | `flags-parsing` |
| P1 | Share ColorMode helper; expose `--color`/`--no-color` on all human formatters | `cli/color` |
| P2 | Extract CLI dispatch out of the fat `browseragent` package / introduce `internal/` | layout |
| P2 | Scripts: prefer `cmd.Debug()` for go/npm/gh orchestration | `cmd-exec` |
| P3 | har-viewer placeholder embed or explicit build docs | `go-embed-assets` |
| P3 | Document firefox-extension-firefox full-embed exception | `go-embed-assets` |
| — | Keep asset hydrate + skill Shape 1 + release dry-run as reference implementations | `go-embed-assets`, `cli/skill-cli`, `cli/dry-run` |

### Concrete first PR sketch (smallest high value)

1. Add `sessionNewHelp` + `parseSessionNewOptions` with less-flags (mirror `parseSessionListOptions`).
2. Unknown flag → fatal; `--help` → session-new-only text.
3. Leave job commands for a second PR with a shared `parseSessionControlOptions` (`--session-id`, `--base-dir`, `--host`, `--server-port`, tab + timeout).

---

## Topic coverage checklist

| Topic | Assessment |
| ----- | ---------- |
| `flags-parsing` | Partial: serve/list good; majority hand-rolled |
| `flags-parsing/subcommand` | Dispatch exists; help/level rule incomplete |
| `flags-parsing/types` | Duration/int often via string parse; weak validation |
| `flags-parsing/cut` / `collect` | N/A (no opaque exec tails / parent flag strip today) |
| `cli/color` | Policy mostly correct where wired; not universal |
| `cli/skill-cli` | Shape 1 done well |
| `cli/streaming` | Adequate for current job UX |
| `cli/dry-run` | Release/scripts good; N/A on main product CLI |
| `cmd-exec` | Scripts mixed; product browser launch OK as raw exec |
| `go-embed-assets` | Reference-quality for browser-agent/trace; har-viewer gap |
| `kool-create` | N/A for retrofit; OK for future greenfield UIs |

---

## Out of scope / not findings

- Extension JS/TS design, CDP protocol details, doctest tree structure  
- Performance of the control server  
- Whether daemon architecture should change  

---

## References (in-repo)

- Skill index: load via `go-best-practice skill --show <topic>`  
- Asset pattern: `docs/assets-hydrate.md`  
- CLI entry: `browseragent/cli.go`, `browseragent/serve.go`, `browseragent/session_list.go`  
- Thin binary: `cmd/browser-agent/main.go`  
- Flag helpers: `browseragent/cli.go` (~1101+), `browseragent/resolve.go`  
- Color: `browseragent/cli_color.go`, `script/internal/clicolor`  
- Dry-run exemplar: `script/github/release/main.go`  
- Embed: `browseragent/embed.go`, `browseragent/assets_*.go`
