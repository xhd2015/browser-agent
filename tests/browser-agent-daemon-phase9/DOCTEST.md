# browser-agent daemon Phase 9 — extension per-session WS + tab routing

Phase 9 changes the **Chrome extension** under `Chrome-Ext-Browser-Agent/public/`
so each session page tab registers with the background worker, opens a
**per-session** WebSocket (`/v1/ws?session=<id>`), and routes CDP jobs to the
correct tab via `payload.session_id`.

| Surface | What is under test |
|---------|-------------------|
| Background WS | Per-session socket URL (`/v1/ws?session=` or session query) |
| Session registry | `register` message + in-memory `sessions` map |
| Content script | On `/go?session=` page: `sendMessage({type:"register", session_id, tabId, windowId})` |
| Job routing | `pickTargetTabId` (or equivalent) scoped by `session_id`; registered tab preferred; URL fallback |
| Manifest | `content_scripts` matches cover loopback go-page host/path |

**No real Chrome.** **No live WebSocket.** Static filesystem/source contract tests
only (same style as `tests/browser-agent-serve-runtime/ext-source/`).

Implementer must also keep embedded extension copy in sync when build copies from
`public/`.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Session Page** (`GET /go?session=<id>`) is opened by the operator. The
**Content Script** runs at `document_start`, reads the session id from the URL,
and sends a **`register`** message to the **Background Service Worker** with
`session_id`, `tabId`, and `windowId`.

**Background Worker** maintains a **`sessions` map** keyed by session id
(`{ws, tabId, windowId}`). On register it connects
`ws://127.0.0.1:PORT/v1/ws?session=<id>` and sends **hello** on that socket.

On **job** messages from the control plane, the background uses
`payload.session_id` to pick the target tab — prefer the registered `tabId` /
`windowId`; fallback to a tab whose URL matches `/go?session=<id>`.

On tab close or navigation away from the session page, the background
**unregisters** the session and closes its WebSocket.

**Test Client** reads extension source files under `ModuleRoot` (no browser).

```text
Content Script on /go?session=S
  -> chrome.runtime.sendMessage({type:"register", session_id:S, tabId, windowId})

Background on register
  -> sessions[S] = {ws, tabId, windowId}
  -> new WebSocket(ws://127.0.0.1:PORT/v1/ws?session=S)
  -> hello

Background on job (payload.session_id = S)
  -> pickTargetTabId(S) — registered tab first, else URL /go?session=S

Tab close / leave go page
  -> unregister(S), close ws
```

## Decision Tree

```
browser-agent-daemon-phase9
├── ext-source/                              [filesystem extension sources]
│   ├── background-per-session-ws/             P1 WS URL includes session query
│   ├── background-session-map/                P2 register + sessions map
│   ├── content-script-register/               P3 content script register message
│   └── job-session-routing/                   P4 pickTargetTabId scoped by session_id
└── manifest/                                [manifest.json content_scripts]
    └── content-script-matches-go/               P5 matches loopback go-page path
```

### Parameter significance (high → low)

1. **Source surface** — background vs content script vs manifest.
2. **Within background** — per-session WS vs session map vs job routing.
3. **Within manifest** — content_scripts `matches` patterns for go page.

## Test Index

| Leaf | Scenario |
|------|----------|
| `ext-source/background-per-session-ws` | (P1) background connects `/v1/ws?session=` (or session query on WS URL) |
| `ext-source/background-session-map` | (P2) handles `register` message; maintains `sessions` map with tab/window ids |
| `ext-source/content-script-register` | (P3) contentScript reads session from URL; `sendMessage` register with `session_id` |
| `ext-source/job-session-routing` | (P4) job tab pick uses `session_id`; registered tab preferred; `/go?session=` fallback |
| `manifest/content-script-matches-go` | (P5) manifest `content_scripts` matches loopback host + `/go` path |

**Leaf count: 5**

## How to Run

```sh
doctest vet ./tests/browser-agent-daemon-phase9
doctest test ./tests/browser-agent-daemon-phase9
# After implementer lands phase 9:
doctest test ./tests/browser-agent-serve-runtime/ext-source/...
doctest test ./tests/browser-agent/...
```

Tree is **RED** until extension sources implement per-session WS + tab routing.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Mode — top-level surface under test.
const (
	ModeExtSource = "ext-source"
	ModeManifest  = "manifest"
)

// ExtSourceTarget for ModeExtSource.
const (
	ExtSrcBackgroundPerSessionWS = "background-per-session-ws"
	ExtSrcBackgroundSessionMap   = "background-session-map"
	ExtSrcContentScriptRegister  = "content-script-register"
	ExtSrcJobSessionRouting      = "job-session-routing"
)

// ManifestProbe for ModeManifest.
const (
	ManifestProbeContentScriptMatchesGo = "content-script-matches-go"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string

	ExtSourceTarget string
	ManifestProbe   string
}

// Response holds filesystem probe outcomes.
type Response struct {
	FoundPaths   []string
	FileExists   bool
	CombinedText string
	FileContents map[string]string
	ErrText      string

	// Manifest-specific
	ManifestText       string
	ContentScriptMatch []string
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
	case ModeExtSource:
		return runExtSource(t, req)
	case ModeManifest:
		return runManifest(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runExtSource(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ExtSourceTarget == "" {
		t.Fatal("ExtSourceTarget must be set")
	}
	root := req.ModuleRoot
	resp := &Response{FileContents: map[string]string{}}

	switch req.ExtSourceTarget {
	case ExtSrcBackgroundPerSessionWS, ExtSrcBackgroundSessionMap, ExtSrcJobSessionRouting:
		candidates := shellBackgroundCandidates(root)
		path, data, ok := firstExistingFile(candidates)
		resp.FileExists = ok
		if ok {
			resp.FoundPaths = []string{path}
			resp.FileContents[path] = string(data)
			resp.CombinedText = string(data)
		} else {
			resp.ErrText = "shell background.js not found under Chrome-Ext-Browser-Agent"
		}
		return resp, nil

	case ExtSrcContentScriptRegister:
		candidates := shellContentScriptCandidates(root)
		path, data, ok := firstExistingFile(candidates)
		resp.FileExists = ok
		if ok {
			resp.FoundPaths = []string{path}
			resp.FileContents[path] = string(data)
			resp.CombinedText = string(data)
		} else {
			resp.ErrText = "contentScript not found under Chrome-Ext-Browser-Agent"
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown ExtSourceTarget %q", req.ExtSourceTarget)
	}
}

func runManifest(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ManifestProbe == "" {
		t.Fatal("ManifestProbe must be set")
	}
	root := req.ModuleRoot
	resp := &Response{FileContents: map[string]string{}}

	candidates := shellManifestCandidates(root)
	path, data, ok := firstExistingFile(candidates)
	resp.FileExists = ok
	if ok {
		resp.FoundPaths = []string{path}
		resp.ManifestText = string(data)
		resp.CombinedText = string(data)
		resp.ContentScriptMatch = extractContentScriptMatches(data)
	} else {
		resp.ErrText = "manifest.json not found under Chrome-Ext-Browser-Agent"
	}
	return resp, nil
}

func shellBackgroundCandidates(root string) []string {
	return []string{
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "public", "background.js"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "background.js"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "src", "background.js"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "build", "background.js"),
	}
}

func shellContentScriptCandidates(root string) []string {
	return []string{
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "public", "contentScript.js"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "public", "content.js"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "contentScript.js"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "src", "contentScript.js"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "build", "contentScript.js"),
	}
}

func shellManifestCandidates(root string) []string {
	return []string{
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "public", "manifest.json"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "manifest.json"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "src", "manifest.json"),
		filepath.Join(root, "Chrome-Ext-Browser-Agent", "build", "manifest.json"),
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

func extractContentScriptMatches(manifestJSON []byte) []string {
	var raw map[string]any
	if err := json.Unmarshal(manifestJSON, &raw); err != nil {
		return nil
	}
	scripts, ok := raw["content_scripts"].([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, s := range scripts {
		m, ok := s.(map[string]any)
		if !ok {
			continue
		}
		matches, ok := m["matches"].([]any)
		if !ok {
			continue
		}
		for _, item := range matches {
			if str, ok := item.(string); ok {
				out = append(out, str)
			}
		}
	}
	return out
}
```