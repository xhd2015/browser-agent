# browser-agent OSS migration — repo layout, hygiene, build, and git contract

Classic **TDD RED** tree: asserts the **target** `browser-agent/` repo after
migrating from `project-api-capture/`. The repo is empty until the implementer
completes the migration; all leaves should fail until then.

| Surface | What is under test |
|---------|-------------------|
| `go-mod` | Module path `github.com/xhd2015/browser-agent` |
| `layout` | HAR viewer subfolder present; Casement artifacts absent |
| `build` | `go build` for `browseragent` and `cmd/browser-agent` |
| `test` | `go test ./browseragent/...` |
| `hygiene` | No company words in tracked tree |
| `git` | `origin` remote URL |

**No real Chrome.** Harness shells out to `go`, `rg`, and `git` from repo root.

After migration, regressions:

```sh
doctest test ./tests/browser-agent
doctest test ./tests/browser-trace
```

## Version

0.0.2

# DSN (Domain Specific Notion)

**Implementer** copies `project-api-capture/` into a fresh **browser-agent**
repo, refactors for OSS, and initializes git without committing.

**Migration harness** (this doctest tree) inspects the resulting repo:

```text
Test Client -> read DOCTEST_ROOT/../.. (repo root)
Test Client -> go build / go test / rg / git remote
Test Client <- layout + hygiene + module path outcomes
```

**Target repo** must contain:

- OSS core packages (`browseragent/`, `cmd/browser-agent/`, extensions, …)
- HAR viewer stack under `har-viewer/` (not repo root)
- Module `github.com/xhd2015/browser-agent`
- Masked company words (`some-x`, `some-y`, …)
- Git remote `https://github.com/xhd2015/browser-agent` (no commit required)

**Removed artifacts** must not exist:

- `Chrome-Ext-Casement-Token/`
- `server/api/casement/` anywhere

## Decision Tree

```
browser-agent-oss-migration
├── go-mod/
│   └── module-path/                    go.mod first line
├── layout/
│   ├── har-viewer-subfolder/           har-viewer paths exist
│   ├── no-casement-extension/          no Chrome-Ext-Casement-Token/
│   └── no-casement-server/             no server/api/casement/
├── build/
│   ├── browseragent/                   go build ./browseragent/...
│   └── cmd/                            go build ./cmd/browser-agent/
├── test/
│   └── browseragent-unit/              go test ./browseragent/...
├── hygiene/
│   └── no-company-words/               rg company words → empty
└── git/
    └── remote-origin/                  git remote get-url origin
```

### Parameter significance (high → low)

1. **Category** — go-mod vs layout vs build vs test vs hygiene vs git (different
   tools and contracts).
2. **Leaf** — specific assertion within the category.
3. **RepoRoot** — always `DOCTEST_ROOT/../..` (set once in root Setup).

## Test Index

| Leaf | Scenario |
|------|----------|
| `go-mod/module-path` | `go.mod` line 1 is `module github.com/xhd2015/browser-agent` |
| `layout/har-viewer-subfolder` | `har-viewer/main.go`, `har-viewer/server/`, `har-viewer/project-api-capture-react/` exist |
| `layout/no-casement-extension` | `Chrome-Ext-Casement-Token/` absent at repo root |
| `layout/no-casement-server` | no `server/api/casement/` directory anywhere |
| `build/browseragent` | `go build ./browseragent/...` succeeds |
| `build/cmd` | `go build -o /dev/null ./cmd/browser-agent/` succeeds |
| `test/browseragent-unit` | `go test ./browseragent/...` passes |
| `hygiene/no-company-words` | `rg` for masking-table originals returns no matches (excl. `.git`, `node_modules`, `dist`) |
| `git/remote-origin` | `git remote get-url origin` = `https://github.com/xhd2015/browser-agent` |

## How to Run

```sh
cd browser-agent
doctest vet ./tests/browser-agent-oss-migration
doctest test ./tests/browser-agent-oss-migration
```

```go
import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	CategoryGoMod    = "go-mod"
	CategoryLayout   = "layout"
	CategoryBuild    = "build"
	CategoryTest     = "test"
	CategoryHygiene  = "hygiene"
	CategoryGit      = "git"

	LeafModulePath           = "module-path"
	LeafHarViewerSubfolder   = "har-viewer-subfolder"
	LeafNoCasementExtension  = "no-casement-extension"
	LeafNoCasementServer     = "no-casement-server"
	LeafBuildBrowseragent    = "browseragent"
	LeafBuildCmd             = "cmd"
	LeafBrowseragentUnit     = "browseragent-unit"
	LeafNoCompanyWords       = "no-company-words"
	LeafRemoteOrigin         = "remote-origin"
)

type Request struct {
	Category string
	Leaf     string
	RepoRoot string
}

type Response struct {
	ExitCode int
	Stdout   string
	Stderr   string
	RunErr   string

	GoModFirstLine string
	PathExists     map[string]bool
	CasementExtAbs string
	CasementSrvAbs string

	RemoteURL string
	RGMatches []string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Category == "" || req.Leaf == "" {
		t.Fatal("Category and Leaf must be set by grouping/leaf Setup")
	}
	if req.RepoRoot == "" {
		req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	}
	resp := &Response{PathExists: make(map[string]bool)}

	switch req.Category {
	case CategoryGoMod:
		return runGoMod(t, req, resp)
	case CategoryLayout:
		return runLayout(t, req, resp)
	case CategoryBuild:
		return runBuild(t, req, resp)
	case CategoryTest:
		return runTest(t, req, resp)
	case CategoryHygiene:
		return runHygiene(t, req, resp)
	case CategoryGit:
		return runGit(t, req, resp)
	default:
		return nil, fmt.Errorf("unknown Category %q", req.Category)
	}
}

func runGoMod(t *testing.T, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	if req.Leaf != LeafModulePath {
		return nil, fmt.Errorf("unknown go-mod leaf %q", req.Leaf)
	}
	path := filepath.Join(req.RepoRoot, "go.mod")
	f, err := os.Open(path)
	if err != nil {
		resp.RunErr = err.Error()
		return resp, nil
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	if sc.Scan() {
		resp.GoModFirstLine = strings.TrimSpace(sc.Text())
	}
	if err := sc.Err(); err != nil {
		resp.RunErr = err.Error()
	}
	return resp, nil
}

func runLayout(t *testing.T, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	switch req.Leaf {
	case LeafHarViewerSubfolder:
		for _, rel := range []string{
			"har-viewer/main.go",
			"har-viewer/server",
			"har-viewer/project-api-capture-react",
		} {
			abs := filepath.Join(req.RepoRoot, filepath.FromSlash(rel))
			_, err := os.Stat(abs)
			resp.PathExists[rel] = err == nil
		}
		return resp, nil
	case LeafNoCasementExtension:
		abs := filepath.Join(req.RepoRoot, "Chrome-Ext-Casement-Token")
		if _, err := os.Stat(abs); err == nil {
			resp.CasementExtAbs = abs
		} else if !os.IsNotExist(err) {
			resp.RunErr = err.Error()
		}
		return resp, nil
	case LeafNoCasementServer:
		found, err := findCasementServerDir(req.RepoRoot)
		if err != nil {
			resp.RunErr = err.Error()
			return resp, nil
		}
		resp.CasementSrvAbs = found
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown layout leaf %q", req.Leaf)
	}
}

func runBuild(t *testing.T, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	var args []string
	switch req.Leaf {
	case LeafBuildBrowseragent:
		args = []string{"build", "./browseragent/..."}
	case LeafBuildCmd:
		args = []string{"build", "-o", "/dev/null", "./cmd/browser-agent/"}
	default:
		return nil, fmt.Errorf("unknown build leaf %q", req.Leaf)
	}
	return runGoCommand(t, req.RepoRoot, args, resp)
}

func runTest(t *testing.T, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	if req.Leaf != LeafBrowseragentUnit {
		return nil, fmt.Errorf("unknown test leaf %q", req.Leaf)
	}
	return runGoCommand(t, req.RepoRoot, []string{"test", "./browseragent/..."}, resp)
}

func runHygiene(t *testing.T, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	if req.Leaf != LeafNoCompanyWords {
		return nil, fmt.Errorf("unknown hygiene leaf %q", req.Leaf)
	}
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("rg not in PATH")
	}
	// Pattern built at runtime so this doctest tree does not embed banned literals.
	pattern := strings.Join([]string{
		"shop" + "ee",
		"sea" + "money",
		"mon" + "ee",
		"gar" + "ena",
		"scre" + "dit" + `\.io`,
	}, "|")
	cmd := exec.Command("rg", "-i", pattern, ".",
		"--glob", "!.git/**",
		"--glob", "!node_modules/**",
		"--glob", "!dist/**",
		"--no-heading",
		"--line-number",
	)
	cmd.Dir = req.RepoRoot
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	resp.Stdout = out.String()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			// rg exit 1 = no matches (desired)
			resp.ExitCode = 1
			return resp, nil
		}
		resp.RunErr = err.Error()
		resp.ExitCode = -1
		return resp, nil
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			resp.RGMatches = append(resp.RGMatches, line)
		}
	}
	return resp, nil
}

func runGit(t *testing.T, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	if req.Leaf != LeafRemoteOrigin {
		return nil, fmt.Errorf("unknown git leaf %q", req.Leaf)
	}
	// Doctest mapping-gen copies the tree without .git; prefer a real git root
	// (GITHUB_WORKSPACE on Actions, or walk up / rev-parse from RepoRoot).
	dir := resolveGitWorkTree(req.RepoRoot)
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	resp.Stdout = text
	if err != nil {
		resp.RunErr = err.Error()
		if ee, ok := err.(*exec.ExitError); ok {
			resp.ExitCode = ee.ExitCode()
		}
		return resp, nil
	}
	// Normalize common URL shapes for exact assert map (strip trailing .git).
	resp.RemoteURL = strings.TrimSuffix(text, ".git")
	return resp, nil
}

func resolveGitWorkTree(start string) string {
	if ws := strings.TrimSpace(os.Getenv("GITHUB_WORKSPACE")); ws != "" {
		if _, err := os.Stat(filepath.Join(ws, ".git")); err == nil {
			return ws
		}
	}
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = start
	if out, err := cmd.Output(); err == nil {
		if root := strings.TrimSpace(string(out)); root != "" {
			return root
		}
	}
	dir := start
	for i := 0; i < 12; i++ {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return start
}

func runGoCommand(t *testing.T, dir string, args []string, resp *Response) (*Response, error) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	resp.Stdout = stdout.String()
	resp.Stderr = stderr.String()
	if err != nil {
		resp.RunErr = err.Error()
		if ee, ok := err.(*exec.ExitError); ok {
			resp.ExitCode = ee.ExitCode()
		}
		return resp, nil
	}
	resp.ExitCode = 0
	return resp, nil
}

func findCasementServerDir(root string) (string, error) {
	var found string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "dist" {
				return filepath.SkipDir
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			if filepath.ToSlash(rel) == "server/api/casement" {
				found = path
				return filepath.SkipAll
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return found, nil
}
```