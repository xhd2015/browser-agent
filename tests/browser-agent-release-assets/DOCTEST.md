# browser-agent release-assets — pack + help + docs (P6–P7)

Classic-TDD tree for the concrete release packaging script and operator docs:

```text
script/github/release-assets/main.go
  go run ./script/github/release-assets [flags]
docs/assets-hydrate.md  (operator cross-doc for pack + --upload)
```

| Surface | What is under test |
|---------|-------------------|
| **Pack** (default, no `--upload`) | Writes three versioned `.tar.gz` archives under `--out` (or a script-created temp dir when `--out` is omitted) from on-disk embeds |
| **Help** | `--help` documents opt-in `--upload` and temp default for `--out` |
| **Docs** (P7) | `docs/assets-hydrate.md` mentions `script/github/release-assets` and `--upload` |

**No real GitHub.** **No `gh`.** **No `--upload` network path.** **No production code in this tree.**

## Mode

**Classic TDD** — pack default-temp-out and help out-default-temp may be **RED** until
implementer lands temp `--out` default. Existing explicit-`--out` pack leaves and
docs may already be GREEN. Do **not** implement production code or edit `docs/`
from this tree.

## Version

0.0.3

# DSN (Domain Specific Notion)

**Release Assets Packer** is a thin module-root script that operators (or CI) run
to materialize the three downloadable hydrate archives that match
`browseragent.AssetReleaseNames(version)` / `EnsureAsset` URL basenames.

**Participants**

- **Operator / CI** — runs the script from the browser-agent module root.
- **Release Assets Script** (`script/github/release-assets`) — packs embed trees into
  `.tar.gz` files; optional `--upload` wraps `gh` (out of scope for these leaves).
- **Embed sources** (read from disk, not runtime `//go:embed` APIs):
  - `browseragent/embedded/session-page`
  - `browseragent/embedded/extension`
  - `browsertrace/embedded/extension`
- **Name Helper** — `browseragent.AssetReleaseNames(version)` yields the three
  basenames, e.g. for `v0.2.0`:
  - `browser-agent_v0.2.0_session-page.tar.gz`
  - `browser-agent_v0.2.0_extension.tar.gz`
  - `browser-trace_v0.2.0_extension.tar.gz`
- **Test Client** — `go run ./script/github/release-assets` from ModuleRoot with
  temp `--out` **or** omitted `--out` (default temp), captures exit/stdout/stderr,
  lists output basenames; docs leaves read `docs/assets-hydrate.md` from ModuleRoot.
  No Chrome.

**Behaviors**

```text
# Pack-only with explicit --out (existing leaves)
go run ./script/github/release-assets --out DIR --version v0.2.0
  -> DIR/browser-agent_v0.2.0_session-page.tar.gz
  -> DIR/browser-agent_v0.2.0_extension.tar.gz
  -> DIR/browser-trace_v0.2.0_extension.tar.gz
  exit 0

# Pack-only with default temp --out (new)
go run ./script/github/release-assets --version v0.2.0
  # (no --out)
  -> os.MkdirTemp("", "browser-agent-release-assets-*")
  -> three .tar.gz under that dir (same names as explicit --out)
  -> stdout includes: out: <abs-path>
  exit 0
  # temp dir need not be deleted on success

# Names must equal package helper
basenames(DIR) == browseragent.AssetReleaseNames("v0.2.0")

# Help
go run ./script/github/release-assets --help
  -> mentions --upload
  -> --out default is temp dir (not "required")
  -> exit 0
  -> trailing newline on help text

# Operator docs (P7)
docs/assets-hydrate.md
  -> mentions script/github/release-assets (pack path)
  -> mentions --upload (gh create / upload --clobber)
```

Pack should set `COPYFILE_DISABLE` / exclude macOS `._*` resource-fork noise
when creating archives (implementer contract; not a separate leaf here).

## Decision Tree

```
browser-agent-release-assets
├── pack/                                      [ModePack — no --upload]
│   ├── writes-three-archives/                   3 non-empty .tar.gz under --out
│   ├── names-match-AssetReleaseNames/           basenames == AssetReleaseNames(ver)
│   └── default-temp-out/                        omit --out → temp dir + out: path + 3 archives
├── help/                                      [ModeHelp]
│   ├── mentions-upload/                         --help mentions --upload
│   └── out-default-temp/                        --help: --out defaults to temp (not required)
└── docs/                                      [ModeDocs — FS read]
    └── assets-hydrate-release/                  assets-hydrate.md: script + --upload
```

### Parameter significance (high → low)

1. **Mode** — pack vs help vs docs (CLI vs filesystem contract).
2. **Within pack** — explicit `--out` vs omitted `--out` (temp default); then
   filesystem existence/size vs set equality with package helper.
3. **Within help** — upload flag vs `--out` default wording.
4. **Within docs** — required tokens in `docs/assets-hydrate.md`.

## Test Index

| Leaf | Scenario |
|------|----------|
| `pack/writes-three-archives` | Pack-only → exit 0; exactly 3 non-empty `.tar.gz` under `--out` with AssetReleaseNames basenames |
| `pack/names-match-AssetReleaseNames` | Out basenames equal `browseragent.AssetReleaseNames(version)` (same multiset) |
| `pack/default-temp-out` | Omit `--out` → exit 0; parse `out: <path>`; 3 non-empty archives matching AssetReleaseNames |
| `help/mentions-upload` | `--help` → exit 0; combined output contains `--upload`; trailing `\n` |
| `help/out-default-temp` | `--help` → mentions temp default for `--out`; does not say required for pack |
| `docs/assets-hydrate-release` | (P7) `docs/assets-hydrate.md` mentions `script/github/release-assets` and `--upload` |

**Leaf count: 6**

## How to Run

```sh
doctest vet ./tests/browser-agent-release-assets
doctest test ./tests/browser-agent-release-assets
# Expect RED on pack/default-temp-out (+ help/out-default-temp) until temp --out default lands
# Existing explicit --out pack leaves should stay GREEN
```

Module: `github.com/xhd2015/browser-agent`.

### Implementer contract (authoritative for GREEN)

```text
// script/github/release-assets/main.go  (package main)
// Proposed behavior sketch (comment at top of main.go):
//   Pack three release asset archives from on-disk embeds for hydrate downloads.
//   Names from browseragent.AssetReleaseNames(version).
//   Default: pack only. --upload opt-in wraps gh (create release if missing,
//   gh release upload --clobber if exists). COPYFILE_DISABLE / exclude ._* .
//
// CLI (from module root):
//   go run ./script/github/release-assets [flags]
//
// Flags:
//   --out DIR         Output directory for archives; when omitted, create via
//                     os.MkdirTemp("", "browser-agent-release-assets-*")
//                     and print "out: <abs-path>" on stdout (preferred last-line
//                     token). Do not require deleting the temp dir on success.
//                     Help: default is temp dir — NOT "required for pack".
//   --version VER     Version string; normalize like AssetReleaseNames
//                     (default: browseragent.ClientVersion() / VERSION.txt)
//   --upload          Opt-in GitHub upload via gh (NOT covered by these leaves)
//   -h, --help        Print usage including --upload and temp --out default;
//                     exit 0; trailing \n
//
// Pack sources (relative to module / process cwd = module root in tests):
//   browseragent/embedded/session-page  -> browser-agent_{ver}_session-page.tar.gz
//   browseragent/embedded/extension     -> browser-agent_{ver}_extension.tar.gz
//   browsertrace/embedded/extension     -> browser-trace_{ver}_extension.tar.gz
//
// Without --upload: write the three archives under --out (or temp default) and exit 0.
// Archive basenames MUST match browseragent.AssetReleaseNames(version) exactly.
// Prefer non-empty valid .tar.gz; exclude AppleDouble ._* entries.
// No real GitHub required for pack/help.
//
// Docs (P7 — docs/assets-hydrate.md):
//   Document operator path:
//     go run ./script/github/release-assets   (pack)
//     --upload                            (gh create if missing; upload --clobber)
//   Prefer real path tokens; no dotted scaffold placeholder IDs.
```

**Non-goals for this leaf set**: `--upload` success/failure paths, real `gh`,
GitHub API mocks, COPYFILE_DISABLE unit tests as separate leaves, auto-delete of
temp out dirs.

```go
import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode — top-level surface under test.
const (
	ModePack = "pack"
	ModeHelp = "help"
	ModeDocs = "docs"
)

// DocsOp — which operator doc probe.
const (
	DocsOpAssetsHydrateRelease = "assets-hydrate-release"
)

// Pin version for deterministic basenames (matches AssetReleaseNames).
const ReleaseVersion = "v0.2.0"

// Script package path relative to ModuleRoot.
const ScriptPkg = "./script/github/release-assets"

// Preferred operator doc path (relative to ModuleRoot).
const AssetsHydrateDocRel = "docs/assets-hydrate.md"

// Mode labels for Request.
type Request struct {
	Mode       string
	ModuleRoot string

	// CLI
	Args []string

	// Pack
	OutDir  string
	Version string
	// PackOmitOut: leaf requests pack without --out; script must create temp dir
	// and print path (prefer "out: <abs-path>" on stdout). Run parses path and
	// lists archives there. Existing leaves leave this false and pass --out.
	PackOmitOut bool

	// Docs
	DocsOp string
}

// Response holds process + pack + docs outcomes.
type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	ErrText  string

	// Pack: basenames found under OutDir (files only, sorted).
	OutBasenames []string
	// Pack: map basename -> size in bytes for non-dir entries.
	OutSizes map[string]int64

	// Docs
	DocsPath string
	DocsText string
	DocsOK   bool
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	if req.ModuleRoot == "" {
		req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	}

	switch req.Mode {
	case ModePack:
		return runPack(t, req)
	case ModeHelp:
		return runHelp(t, req)
	case ModeDocs:
		return runDocs(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runPack(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	ver := req.Version
	if ver == "" {
		ver = ReleaseVersion
	}

	args := append([]string{}, req.Args...)
	if req.PackOmitOut {
		// Default-temp-out path: never inject --out; leaf pins Args to --version only.
		if len(args) == 0 {
			args = []string{"--version", ver}
		}
	} else {
		if req.OutDir == "" {
			t.Fatal("OutDir must be set for pack (or set PackOmitOut for default temp)")
		}
		if len(args) == 0 {
			args = []string{
				"--out", req.OutDir,
				"--version", ver,
			}
		}
	}

	resp, err := runScript(t, req.ModuleRoot, args)
	if err != nil {
		return resp, err
	}

	// Resolve out directory: explicit OutDir, else parse from stdout.
	outDir := req.OutDir
	if req.PackOmitOut || outDir == "" {
		if p, ok := parseOutDirFromStdout(resp.Stdout); ok {
			outDir = p
			req.OutDir = p
		} else if p, ok := parseOutDirFromStdout(resp.Stdout + "\n" + resp.Stderr); ok {
			outDir = p
			req.OutDir = p
		}
	}

	// List archives under resolved OutDir even when exit != 0 (partial writes).
	if outDir != "" {
		resp.OutBasenames, resp.OutSizes = listOutArchives(t, outDir)
	}
	return resp, nil
}

// parseOutDirFromStdout extracts the pack output directory from script logs.
// Preferred: a line "out: <path>" (case-sensitive token "out:", path trimmed).
// Fallback: a line containing "packing into" followed by a path token.
func parseOutDirFromStdout(stdout string) (string, bool) {
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "out:") {
			path := strings.TrimSpace(strings.TrimPrefix(trim, "out:"))
			if path != "" {
				return path, true
			}
		}
	}
	// Fallback: packing into <path>
	lowerInto := "packing into"
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		idx := strings.Index(strings.ToLower(trim), lowerInto)
		if idx < 0 {
			continue
		}
		rest := strings.TrimSpace(trim[idx+len(lowerInto):])
		// Drop trailing punctuation / size notes.
		if rest == "" {
			continue
		}
		// First path-like token (stop at whitespace if extra text).
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}
		path := strings.Trim(fields[0], `"'`)
		if path != "" {
			return path, true
		}
	}
	return "", false
}

func runHelp(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	args := req.Args
	if len(args) == 0 {
		args = []string{"--help"}
	}
	return runScript(t, req.ModuleRoot, args)
}

func runScript(t *testing.T, moduleRoot string, scriptArgs []string) (*Response, error) {
	t.Helper()
	// go run ./script/github/release-assets <args...>
	goArgs := append([]string{"run", ScriptPkg}, scriptArgs...)
	cmd := exec.Command("go", goArgs...)
	cmd.Dir = moduleRoot
	// Isolation: do not inherit upload tokens; pack must not need network.
	cmd.Env = append(os.Environ(),
		"COPYFILE_DISABLE=1",
		// Avoid accidental gh auth noise if implementer shells out unexpectedly.
		"GH_TOKEN=",
		"GITHUB_TOKEN=",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Bound runtime (go run may compile).
	timer := time.AfterFunc(3*time.Minute, func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	})
	defer timer.Stop()

	runErr := cmd.Run()
	resp := &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
	}
	if runErr != nil {
		resp.ErrText = runErr.Error()
		if ee, ok := runErr.(*exec.ExitError); ok {
			resp.ExitCode = ee.ExitCode()
		} else {
			// e.g. script package missing / go not found — treat as hard fail for assert.
			resp.ExitCode = -1
		}
	}
	return resp, nil
}

func listOutArchives(t *testing.T, outDir string) ([]string, map[string]int64) {
	t.Helper()
	sizes := map[string]int64{}
	entries, err := os.ReadDir(outDir)
	if err != nil {
		// OutDir missing or unreadable — empty listing; assert will fail.
		return nil, sizes
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// Ignore junk (AppleDouble, .DS_Store) if any leaked into out dir.
		if strings.HasPrefix(name, "._") || name == ".DS_Store" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		names = append(names, name)
		sizes[name] = info.Size()
	}
	sort.Strings(names)
	return names, sizes
}

// expectedReleaseNames returns AssetReleaseNames for the pinned version.
// Used by asserts; also documents the package dependency for implementers.
func expectedReleaseNames(version string) []string {
	if version == "" {
		version = ReleaseVersion
	}
	return browseragent.AssetReleaseNames(version)
}

func runDocs(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.DocsOp == "" {
		t.Fatal("DocsOp must be set by leaf Setup")
	}
	resp := &Response{ExitCode: 0}
	switch req.DocsOp {
	case DocsOpAssetsHydrateRelease:
		path := filepath.Join(req.ModuleRoot, AssetsHydrateDocRel)
		resp.DocsPath = path
		b, err := os.ReadFile(path)
		if err != nil {
			resp.DocsOK = false
			resp.ErrText = err.Error()
			// File missing is a soft response for Classic TDD asserts (not transport err).
			return resp, nil
		}
		resp.DocsOK = true
		resp.DocsText = string(b)
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown DocsOp %q", req.DocsOp)
	}
}
```
