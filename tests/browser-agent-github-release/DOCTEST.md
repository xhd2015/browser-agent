# browser-agent GitHub release CLI + install.sh

Classic-TDD surface for fat binary release (doctest-style) and curl installer.

```text
script/github/release/main.go
  go run ./script/github/release [--dry-run] [--skip-prebuild] [--skip-hydrate]
install.sh
  curl …/install.sh | bash
```

| Surface | What is under test |
|---------|-------------------|
| **Help** | `--help` documents dry-run / skip flags / credentials / install.sh |
| **Dry-run** | `--dry-run` exit 0; plans binary names `browser-agent-v*-{os}-{arch}`; soft-warns dirty git/creds |
| **install.sh** | Repo root script exists; download formula tokens match release asset names |

**No real GitHub upload.** **No multi-arch live build.** **No network download.**

## Version

0.0.1

# DSN

**Release CLI** builds multi-platform fat `browser-agent` binaries and uploads
them (plus hydrate archives) to GitHub Releases — same pattern as
`github.com/xhd2015/doctest/script/github/release`.

**install.sh** downloads `browser-agent-{tag}-{os}-{arch}` from the latest
(or pinned) release into `$GOPATH/bin` or `/usr/local/bin`.

## Decision Tree

```
browser-agent-github-release
├── help/
│   └── mentions-dry-run/              --help mentions --dry-run + install.sh
├── dry-run/
│   └── plans-binaries/                --dry-run prints would build: browser-agent-v*…
└── install-sh/
    └── script-present/                install.sh exists; formula tokens
```

## How to Run

```sh
doctest vet ./tests/browser-agent-github-release
doctest test ./tests/browser-agent-github-release
```

Module: `github.com/xhd2015/browser-agent`.

```go
import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	ModeHelp      = "help"
	ModeDryRun    = "dry-run"
	ModeInstallSH = "install-sh"
)

const ScriptPkg = "./script/github/release"

type Request struct {
	Mode       string
	ModuleRoot string
	Args       []string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	ErrText  string

	// install-sh
	InstallSHPath string
	InstallSHText string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	if req.ModuleRoot == "" {
		req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	}

	switch req.Mode {
	case ModeHelp, ModeDryRun:
		return runCLI(t, req)
	case ModeInstallSH:
		return runInstallSH(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runCLI(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	args := append([]string{"run", ScriptPkg}, req.Args...)
	ctxTO := 3 * time.Minute
	cmd := exec.Command("go", args...)
	cmd.Dir = req.ModuleRoot
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	done := make(chan error, 1)
	go func() { done <- cmd.Run() }()

	var runErr error
	select {
	case runErr = <-done:
	case <-time.After(ctxTO):
		_ = cmd.Process.Kill()
		runErr = fmt.Errorf("go run timed out after %s", ctxTO)
	}

	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			resp.ExitCode = ee.ExitCode()
		} else {
			resp.ExitCode = 1
			resp.ErrText = runErr.Error()
		}
	} else {
		resp.ExitCode = 0
	}
	// Transport success: CLI may exit non-zero intentionally — Assert checks ExitCode.
	return resp, nil
}

func runInstallSH(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	p := filepath.Join(req.ModuleRoot, "install.sh")
	data, err := os.ReadFile(p)
	resp := &Response{InstallSHPath: p}
	if err != nil {
		resp.ExitCode = 1
		resp.ErrText = err.Error()
		return resp, nil
	}
	resp.InstallSHText = string(data)
	resp.ExitCode = 0
	return resp, nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport/setup error: %v", err)
	}
}

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode != 0 {
		t.Fatalf("exit=%d stdout=%s stderr=%s err=%s",
			resp.ExitCode, truncate(resp.Stdout, 400), truncate(resp.Stderr, 400), resp.ErrText)
	}
}

func combinedOut(resp *Response) string {
	if resp == nil {
		return ""
	}
	return resp.Stdout + resp.Stderr
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
```
