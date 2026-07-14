# canonical-path — EnsureCanonicalExtension layout (nested)

Nested doctest for **`EnsureCanonicalExtension`** and **`DefaultExtensionInstallLayout`**.
Isolated from parent tree because these symbols are not yet in `browseragent` (compile RED
until implementer lands canonical extract APIs).

## Version

0.0.2

# DSN (Domain Specific Notion)

**Canonical Extension Layout** under `TestHome`:

```text
~/.browser-agent/managed-chrome/extensions/browser-agent/<version>/
```

`EnsureCanonicalExtension()` writes embedded MV3 once per version; second call is idempotent.

## Decision Tree

```
canonical-path
├── extracts-under-browser-agent/    manifest.json under browser-agent/{ver}/
└── idempotent-same-version/         second call same path + version
```

## Test Index

| Leaf | Scenario |
|------|----------|
| `extracts-under-browser-agent` | First extract → `extensions/browser-agent/{ver}/manifest.json` |
| `idempotent-same-version` | Second extract → same path + version |

**Leaf count: 2**

## How to Run

```sh
doctest vet ./tests/browser-agent-extension-install-workflow/canonical-path
doctest test ./tests/browser-agent-extension-install-workflow/canonical-path
```

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/browser-agent/browseragent"
)

const (
	CanonicalPathOpExtractsUnderBrowserAgent = "extracts-under-browser-agent"
	CanonicalPathOpIdempotentSameVersion     = "idempotent-same-version"
)

type Request struct {
	ModuleRoot string
	TestHome   string
	CanonicalPathOp string
}

type Response struct {
	ExtensionPath  string
	ExtensionVer   string
	ExtensionPath2 string
	ExtensionVer2  string
	ManifestPath   string
	CanonicalExtensionsDir string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req == nil {
		t.Fatal("req is nil")
	}
	if req.CanonicalPathOp == "" {
		t.Fatal("CanonicalPathOp must be set")
	}
	if req.TestHome != "" {
		t.Setenv("HOME", req.TestHome)
	}

	resp := &Response{}
	layout, err := browseragent.DefaultExtensionInstallLayout()
	if err != nil {
		return resp, err
	}
	resp.CanonicalExtensionsDir = layout.BrowserAgentExtensionsDir

	path1, ver1, err := browseragent.EnsureCanonicalExtension()
	if err != nil {
		return resp, err
	}
	resp.ExtensionPath = path1
	resp.ExtensionVer = ver1
	resp.ManifestPath = filepath.Join(path1, "manifest.json")

	switch req.CanonicalPathOp {
	case CanonicalPathOpExtractsUnderBrowserAgent:
		return resp, nil
	case CanonicalPathOpIdempotentSameVersion:
		path2, ver2, err2 := browseragent.EnsureCanonicalExtension()
		if err2 != nil {
			return resp, err2
		}
		resp.ExtensionPath2 = path2
		resp.ExtensionVer2 = ver2
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown CanonicalPathOp %q", req.CanonicalPathOp)
	}
}
```