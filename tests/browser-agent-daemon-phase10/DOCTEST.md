# browser-agent daemon Phase 10 — docs & CLI help polish

Phase 10 updates **operator docs** and **CLI help strings** for the daemon
workflow:

- `serve` — blocking daemon host
- `serve --status` — read-only status probe
- `serve --kill-existing` — shutdown existing daemon before start
- `session new [--session-id]` — ensure daemon + create session + open Chrome
- Deprecation note for `serve --session-id`
- `cmd/browser-agent/SKILL.md` — workflow: `serve` or `session new` →
  `session info|eval|…`

**No daemon spawn. No Chrome.** Pure text contract tests reading `cli.go` help
via `HandleCLI --help` and `SKILL.md` from the module tree.

| Surface | What is under test |
|---------|-------------------|
| `briefUsage` / `fullHelp` | Top-level + serve + session help markers |
| `SKILL.md` | Operator workflow + `serve --status` + deprecation |

Depends on Phases 5–8 (daemon serve, status, kill-existing, session new).

## Version

0.0.2

# DSN (Domain Specific Notion)

**Operator CLI help** (`browseragent/cli.go`):

```text
briefUsage  — top-level commands; must name session new; serve as blocking daemon host
fullHelp    — serve flags: --status (read-only), --kill-existing; deprecate serve --session-id
              session new flags: --session-id auto-generate when omitted
```

**Agent skill doc** (`cmd/browser-agent/SKILL.md`, mirrored in package embed):

```text
Bootstrap: browser-agent serve   OR   browser-agent session new [--session-id]
Probe:     browser-agent serve --status
Work:      browser-agent session info|eval|run|logs|screenshot|cdp …
Legacy:    serve --session-id deprecated (prefer session new or plain serve)
```

**Test Client** calls `HandleCLI` with stdout/stderr buffers or reads SKILL.md
from `ModuleRoot` (no binary shell-out).

## Decision Tree

```
browser-agent-daemon-phase10
├── cli-help/                                    [HandleCLI help text]
│   ├── top-level-mentions-session-new/            brief + full: session new, blocking serve
│   ├── serve-mentions-status-kill/                serve flags: --status, --kill-existing, deprecation
│   └── session-help-mentions-new/                 session help: session new + auto-generate
└── skill-md/                                    [SKILL.md filesystem]
    └── workflow-session-new/                      session new, serve --status, session cmds, deprecation
```

### Parameter significance (high → low)

1. **Surface** — CLI help strings vs SKILL.md prose.
2. **Within CLI** — top-level vs serve vs session subcommand help.
3. **Within SKILL** — bootstrap workflow vs status probe vs deprecation.

## Test Index

| Leaf | Scenario |
|------|----------|
| `cli-help/top-level-mentions-session-new` | `--help` + briefUsage mention `session new`; serve blocking daemon host |
| `cli-help/serve-mentions-status-kill` | `serve --help` documents `--status`, `--kill-existing`, deprecates `--session-id` |
| `cli-help/session-help-mentions-new` | `session --help` documents `session new` + auto-generate `--session-id` |
| `skill-md/workflow-session-new` | SKILL.md: `session new`, `serve --status`, session side cmds, deprecation |

**Leaf count: 4**

## How to Run

```sh
doctest vet ./tests/browser-agent-daemon-phase10
doctest test ./tests/browser-agent-daemon-phase10
# After implementer lands phase 10:
doctest test ./tests/browser-agent-daemon-phase8
doctest test ./tests/browser-agent-vite-skill/...
```

Tree is **RED** until `cli.go` help strings and `SKILL.md` are polished.

```go
import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode — top-level surface under test.
const (
	ModeCLIHelp = "cli-help"
	ModeSkillMD = "skill-md"
)

// CLIHelpProbe — help text contracts.
const (
	CLIHelpProbeTopLevelSessionNew = "top-level-mentions-session-new"
	CLIHelpProbeServeStatusKill    = "serve-mentions-status-kill"
	CLIHelpProbeSessionHelpNew     = "session-help-mentions-new"
)

// SkillMDProbe — SKILL.md contracts.
const (
	SkillMDProbeWorkflowSessionNew = "workflow-session-new"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string

	CLIHelpProbe      string
	CLIArgs           []string
	CaptureBriefUsage bool
	BriefUsageArgs    []string

	SkillMDProbe string
}

// Response holds help / skill probe outcomes.
type Response struct {
	HelpText       string
	BriefUsageText string
	Stdout         string
	Stderr         string
	CLIErr         string

	SkillFileExists bool
	SkillText       string
	SkillPath       string
	SkillPathsTried []string
	ErrText         string
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
	case ModeCLIHelp:
		return runCLIHelp(t, req)
	case ModeSkillMD:
		return runSkillMD(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runCLIHelp(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.CLIHelpProbe == "" {
		t.Fatal("CLIHelpProbe must be set by leaf Setup")
	}
	args := req.CLIArgs
	if args == nil {
		switch req.CLIHelpProbe {
		case CLIHelpProbeTopLevelSessionNew:
			args = []string{"--help"}
		case CLIHelpProbeServeStatusKill:
			args = []string{"serve", "--help"}
		case CLIHelpProbeSessionHelpNew:
			args = []string{"session", "--help"}
		default:
			args = []string{"--help"}
		}
	}

	resp := &Response{}
	var stdout, stderr bytes.Buffer
	err := browseragent.HandleCLI(args, map[string]string{}, &stdout, &stderr)
	if err != nil {
		resp.CLIErr = err.Error()
	} else {
		resp.CLIErr = ""
	}
	resp.Stdout = stdout.String()
	resp.Stderr = stderr.String()
	resp.HelpText = resp.Stdout

	if req.CaptureBriefUsage {
		briefArgs := req.BriefUsageArgs
		if briefArgs == nil {
			briefArgs = []string{}
		}
		var briefOut, briefErr bytes.Buffer
		_ = browseragent.HandleCLI(briefArgs, map[string]string{}, &briefOut, &briefErr)
		// Bare session prints briefUsage on stderr; top-level bare uses stdout.
		if briefOut.Len() == 0 && briefErr.Len() > 0 {
			resp.BriefUsageText = briefErr.String()
		} else if briefOut.Len() > 0 && briefErr.Len() > 0 {
			resp.BriefUsageText = briefOut.String() + briefErr.String()
		} else {
			resp.BriefUsageText = briefOut.String()
		}
	}

	return resp, nil
}

func runSkillMD(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SkillMDProbe == "" {
		t.Fatal("SkillMDProbe must be set by leaf Setup")
	}
	root := req.ModuleRoot
	resp := &Response{}

	candidates := skillMDCandidates(root)
	resp.SkillPathsTried = candidates
	path, data, ok := firstExistingFile(candidates)
	resp.SkillFileExists = ok
	if ok {
		resp.SkillPath = path
		resp.SkillText = string(data)
	} else {
		resp.ErrText = "SKILL.md not found under cmd/browser-agent or browseragent"
	}
	return resp, nil
}

func skillMDCandidates(root string) []string {
	return []string{
		filepath.Join(root, "cmd", "browser-agent", "SKILL.md"),
		filepath.Join(root, "browseragent", "SKILL.md"),
	}
}

func firstExistingFile(paths []string) (string, []byte, bool) {
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			return p, data, true
		}
	}
	return "", nil, false
}

// Optional source probe (not used by default leaves): extract const blocks from cli.go.
func readCLIGoHelpSource(moduleRoot string) (brief, full string, err error) {
	path := filepath.Join(moduleRoot, "browseragent", "cli.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	text := string(data)
	brief, ok := extractGoConstString(text, "briefUsage")
	if !ok {
		return "", "", fmt.Errorf("briefUsage const not found in %s", path)
	}
	full, ok = extractGoConstString(text, "fullHelp")
	if !ok {
		return "", "", fmt.Errorf("fullHelp const not found in %s", path)
	}
	return brief, full, nil
}

func extractGoConstString(src, name string) (string, bool) {
	marker := "const " + name + " = `"
	idx := strings.Index(src, marker)
	if idx < 0 {
		return "", false
	}
	start := idx + len(marker)
	end := strings.Index(src[start:], "`")
	if end < 0 {
		return "", false
	}
	return src[start : start+end], true
}

// Silence unused import when helper is not referenced by generated test glue.
var _ = io.Discard
```