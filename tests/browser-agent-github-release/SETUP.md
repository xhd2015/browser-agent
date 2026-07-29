# Scenario

**Feature**: GitHub release CLI (help, dry-run) and install.sh formula

```
go run ./script/github/release --help
go run ./script/github/release --dry-run
install.sh present at module root
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Tree root `tests/browser-agent-github-release/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- No network; dry-run must not require credentials or clean git.

## Steps

1. Resolve ModuleRoot from DOCTEST_ROOT.
2. Leave Mode empty at root (grouping/leaf sets it).

## Context

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	if st, err := os.Stat(req.ModuleRoot); err != nil || !st.IsDir() {
		t.Fatalf("ModuleRoot %s not a directory: %v", req.ModuleRoot, err)
	}
	return nil
}
```
