# browser-agent Vite session-page embed + skill packaging

Exercises **committed Vite session-page production fixtures** embedded into
package `browseragent`, served for session SPA routes, plus **skill CLI**
packaging parallel to `browser-trace`.

| Surface | What is under test |
|---------|-------------------|
| Embed FS | `//go:embed embedded/session-page/**` index HTML has root mount |
| HTTP SPA | `GET /go` and `GET /` inject boot/product markers + session poll hooks |
| Install UX | Not-connected install markers still present on SPA HTML |
| Static asset | Fixture asset under embed served with HTTP 200 |
| Skill CLI | `HandleCLI` `skill --list` / `skill --show` / bare skill |
| Boot JSON | Pure `FormatSessionBootJSON(sessionID)` helper |

**No npm. No real Chrome. No real agent-run.** CI uses a **committed** mini
fixture under `browseragent/embedded/session-page/` (optional operator bundle
script `script/browser-agent/bundle` is out of scope for this tree).

**Sealed** prior trees (do **not** modify; must stay GREEN):

```sh
doctest test ./tests/browser-agent/...
doctest test ./tests/browser-agent-cli-react/...
doctest test ./tests/browser-agent-serve-runtime/...
doctest test ./tests/browser-agent-cdp-jobs/...
```

Prior SPA leaves in `browser-agent-cli-react/spa-embed` stay GREEN — keep
root/`/v1/session`/43761/`browser-agent`/install markers.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Operator / Agent** uses **`browser-agent`** (package `HandleCLI` preferred
over binary shell-out). This cycle adds:

1. **Session-page embed** — Vite (or minimal fixture) assets staged at:

```text
browseragent/embedded/session-page/
  index.html                 # entry (session-page.html also acceptable)
  assets/session-page.js     # tiny committed fixture JS (required for B1)
  assets/…                   # optional CSS/other
```

`//go:embed embedded/session-page/**` ships the tree in the binary/package.

2. **Control Server** (`browseragent.Run`, `NoOpenChrome`, `NoAgentRun`) serves:

- `GET /go` and `GET /` → session SPA HTML from embed (prefer real embed over
  pure Go HTML string when embed is present)
- Static files under `/assets/…` (or path-relative to embed) when fixture has them
- Boot config injected so React can read session id, product id
  `browser-agent`, control port **43761**, poll `/v1/session`:

```text
<script type="application/json" id="browser-agent-boot">
  {"session_id":"…","product":"browser-agent","control_port":43761}
</script>
# and/or data-session-id / data-product / data-control-port / window.__BROWSER_AGENT
```

HTML must still satisfy prior SPA markers (root mount, `/v1/session`, `43761`,
`browser-agent`, install markers) — either in embedded index or via injection.

3. **Skill packaging** (Shape 1, like browser-trace via `skillcmd.SingleSkill`):

```text
cmd/browser-agent/SKILL.md   # name: browser-agent (//go:embed in cmd or package)
```

```text
browser-agent skill --list               → "browser-agent\n"
browser-agent skill --show               → skill body + trailing \n
browser-agent skill                      → help or skillcmd-consistent error
browser-agent skill --install …          → install (not asserted in this tree)
```

`--show` body markers: product `browser-agent`, env `BROWSER_AGENT_SESSION_ID`,
nested `session` side-command parent,
side commands (`eval` / `info` / …), control port **43761**.

4. **Boot JSON pure helper** (no server):

```text
FormatSessionBootJSON(sessionID) → JSON with session_id, product, control_port
```

**Test Client** in this tree:

- Embed leaf reads package embed FS / ReadFile helpers (no HTTP)
- HTTP leaves start `browseragent.Run` and GET paths
- Skill leaves call `HandleCLI` with injectable env + buffers
- Boot leaf calls pure helper only

## Decision Tree

```
browser-agent-vite-skill
├── embed-session-page/                        [ModeEmbedFS]
│   └── index-root-mount/                        A1 index HTML non-empty + root mount
├── http-session-page/                         [ModeHTTP]
│   ├── go-boot-markers/                         A2 GET /go → session + boot + poll + port
│   ├── root-product-marker/                     A3 GET /  → product browser-agent
│   ├── install-markers-not-connected/           A4 install UX markers (no WS)
│   └── static-asset/                            B1 GET fixture asset → 200
├── skill-cli/                                 [ModeSkill]
│   ├── list/                                    C1 skill --list → browser-agent\n
│   ├── show/                                    C2 skill --show → markers +\n
│   └── bare-help/                               C3 skill (no flags) → help|error
└── boot-json/                                 [ModeBootJSON]
    └── format-session-boot/                     D1 FormatSessionBootJSON fields
```

### Parameter significance (high → low)

1. **Surface / Mode** — embed FS vs live HTTP vs skill CLI vs pure boot helper
   (different `Run` branches and contracts).
2. **Within HTTP** — path / probe (`/go` boot, `/` product, install markers,
   static asset).
3. **Within skill** — action (`--list` / `--show` / bare).
4. **Session id value** — only for boot JSON / HTTP boot injection (leaf sets).

## Test Index

| Leaf | Scenario |
|------|----------|
| `embed-session-page/index-root-mount` | (A1) Embed index HTML non-empty; root mount `id="root"` or `data-browser-agent-root` |
| `http-session-page/go-boot-markers` | (A2) GET `/go` after Run: session id + boot/product + `/v1/session` + `43761` |
| `http-session-page/root-product-marker` | (A3) GET `/` (or redirect→go body): product `browser-agent` present |
| `http-session-page/install-markers-not-connected` | (A4) No extension: `chrome://extensions` or Load unpacked / install marker |
| `http-session-page/static-asset` | (B1) Fixture asset GET 200 non-empty (prefer `/assets/session-page.js`) |
| `skill-cli/list` | (C1) `skill --list` → stdout contains `browser-agent`; trailing `\n`; nil CLI err |
| `skill-cli/show` | (C2) `skill --show` → `BROWSER_AGENT_SESSION_ID`, `session`, `eval`, `43761`, `browser-agent`; trailing `\n` |
| `skill-cli/bare-help` | (C3) `skill` alone → help text and/or non-nil err consistent with skillcmd; no hang |
| `boot-json/format-session-boot` | (D1) `FormatSessionBootJSON` → `session_id`, `product=browser-agent`, `control_port=43761` |

**Leaf count: 9**

## How to Run

```sh
cd project-api-capture
doctest vet ./tests/browser-agent-vite-skill
doctest test ./tests/browser-agent-vite-skill/...

# regressions — sealed prior browser-agent trees
doctest test ./tests/browser-agent/...
doctest test ./tests/browser-agent-cli-react/...
doctest test ./tests/browser-agent-serve-runtime/...
doctest test ./tests/browser-agent-cdp-jobs/...
```

Module: `github.com/xhd2015/browser-agent`.  
Package under test: `github.com/xhd2015/browser-agent/browseragent`
(and thin `cmd/browser-agent` for skill embed if content lives only on cmd).

### Implementer contract (authoritative for GREEN)

#### Session-page embed fixture (committed; no npm in CI)

```text
browseragent/embedded/session-page/index.html
browseragent/embedded/session-page/assets/session-page.js
```

- `index.html` (or `session-page.html`) includes a root mount marker
  (`id="root"` and/or `data-browser-agent-root`) so A1 passes even before
  server injection.
- Tiny JS fixture is enough for B1; production Vite output may replace later
  via `script/browser-agent/bundle` (not run by tests).

#### Package APIs

```go
// Session-page embed FS (required for A1):
//
//   //go:embed embedded/session-page/**
//   var sessionPage embed.FS  // unexported ok
//
//   func SessionPageFS() fs.FS  // required export; returns embed root
//
// Root file names accepted by harness: "index.html" or "session-page.html"
// (with or without "embedded/session-page/" prefix depending on how embed is
// rooted — harness probes both).

// Boot helper (D1) — required export:
func FormatSessionBootJSON(sessionID string) string
// Must parse as JSON object with at least:
//   session_id   (string, equals input)
//   product      ("browser-agent")
//   control_port (43761 as number; string "43761" also accepted)

// Existing (already in package):
func Run(ctx context.Context, cfg Config) (…, error)
func HandleCLI(args []string, env map[string]string, stdout, stderr io.Writer) error
```
#### HTTP serve

- Prefer serving embed HTML for `GET /go` and `GET /` when embed present.
- Inject boot (script `#browser-agent-boot` and/or data attrs / `__BROWSER_AGENT`)
  with live session id, product `browser-agent`, port `43761`.
- Preserve install markers when not connected.
- Serve fixture static asset(s) at `/assets/session-page.js` (or equivalent
  under `/assets/`).

#### Skill CLI via HandleCLI

```text
HandleCLI([]string{"skill", "--list"}, …)
HandleCLI([]string{"skill", "--show"}, …)
HandleCLI([]string{"skill"}, …)
```

- Output must appear on the **stdout/stderr writers** passed to HandleCLI
  (do not only write to process `os.Stdout` without capturing into those
  writers — tests do not shell out to the binary).
- `--list`: skill name `browser-agent` + trailing `\n`; nil error.
- `--show`: body includes markers above; trailing `\n`; nil error.
- bare `skill`: skillcmd-consistent — either print help (nil or non-nil err)
  or non-nil err mentioning `--show` / `--list` / `--install` / `--help`;
  must not hang / serve.
- Content source: `//go:embed` of `cmd/browser-agent/SKILL.md` or package-level
  skill content; name `browser-agent`. Pattern: `github.com/xhd2015/skills/skillcmd`.

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode values — top-level surface under test.
const (
	ModeEmbedFS  = "embed-fs"
	ModeHTTP     = "http"
	ModeSkill    = "skill"
	ModeBootJSON = "boot-json"
)

// HTTPProbe selects the HTTP leaf contract under ModeHTTP.
const (
	HTTPProbeGoBoot  = "go-boot"
	HTTPProbeRoot    = "root-product"
	HTTPProbeInstall = "install-markers"
	HTTPProbeAsset   = "static-asset"
)

// SkillAction under ModeSkill.
const (
	SkillActionList = "list"
	SkillActionShow = "show"
	SkillActionBare = "bare"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	// Mode selects the surface Run executes.
	Mode string

	// ModuleRoot is project-api-capture module directory.
	// Root Setup sets from DOCTEST_ROOT/../..
	ModuleRoot string

	// BaseDir is temp parent for serve state.
	BaseDir string

	// Session / server (HTTP modes)
	Addr         string
	SessionID    string
	NoOpenChrome bool
	NoAgentRun   bool
	ReadyTimeout time.Duration

	// HTTPProbe: go-boot | root-product | install-markers | static-asset
	HTTPProbe string
	// AssetPath is URL path for static-asset (default "/assets/session-page.js").
	AssetPath string

	// Skill (ModeSkill)
	SkillAction     string
	CLIArgs         []string
	CLIEnv          map[string]string
	MaxDispatchWait time.Duration

	// Boot JSON (ModeBootJSON)
	BootSessionID string
}

// Response holds outcomes for all modes (fields used per mode).
type Response struct {
	// Embed FS
	EmbedPath    string
	EmbedHTML    string
	EmbedListing []string

	// HTTP
	StatusCode    int
	ContentType   string
	Body          []byte
	BodyString    string
	BaseURL       string
	ProbeURL      string
	RealSessionID string
	RedirectURL   string

	// CLI / skill
	Stdout           string
	Stderr           string
	ExitCode         int
	ErrText          string
	CLIErr           string
	DispatchTimedOut bool

	// Boot JSON
	BootJSON        string
	BootSessionID   string
	BootProduct     string
	BootControlPort int
	BootPortStr     string
	BootParseOK     bool
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
	case ModeEmbedFS:
		return runEmbedFS(t, req)
	case ModeHTTP:
		return runHTTP(t, req)
	case ModeSkill:
		return runSkill(t, req)
	case ModeBootJSON:
		return runBootJSON(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

// --- Embed FS ---

func runEmbedFS(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	html, pathName, listing, err := readEmbeddedSessionIndex()
	resp := &Response{
		EmbedPath:    pathName,
		EmbedHTML:    html,
		EmbedListing: listing,
		BodyString:   html,
	}
	if err != nil {
		resp.ErrText = err.Error()
		return resp, err
	}
	return resp, nil
}

func readEmbeddedSessionIndex() (html, pathName string, listing []string, err error) {
	fsys := browseragent.SessionPageFS()
	if fsys == nil {
		return "", "", nil, fmt.Errorf("browseragent.SessionPageFS() returned nil")
	}
	root := detectSessionPageRoot(fsys)

	_ = fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}
		rel := filepath.ToSlash(p)
		if root != "." && root != "" {
			if r, e := filepath.Rel(filepath.FromSlash(root), filepath.FromSlash(p)); e == nil {
				rel = filepath.ToSlash(r)
			}
		}
		listing = append(listing, rel)
		return nil
	})

	candidates := []string{
		joinEmbed(root, "index.html"),
		joinEmbed(root, "session-page.html"),
		"index.html",
		"session-page.html",
	}
	for _, rel := range listing {
		base := path.Base(rel)
		if base == "index.html" || base == "session-page.html" {
			candidates = append(candidates, joinEmbed(root, rel), rel)
		}
	}

	var last error
	seen := map[string]bool{}
	for _, c := range candidates {
		c = filepath.ToSlash(c)
		if c == "" || seen[c] {
			continue
		}
		seen[c] = true
		data, rerr := fs.ReadFile(fsys, c)
		if rerr != nil {
			last = rerr
			continue
		}
		return string(data), c, listing, nil
	}
	if last == nil {
		last = fmt.Errorf("no session-page index under embed (listing=%v)", listing)
	}
	return "", "", listing, last
}

func joinEmbed(root, name string) string {
	name = strings.TrimPrefix(filepath.ToSlash(name), "/")
	if root == "" || root == "." {
		return name
	}
	return filepath.ToSlash(path.Join(filepath.ToSlash(root), name))
}

func detectSessionPageRoot(fsys fs.FS) string {
	for _, root := range []string{".", "embedded/session-page", "session-page"} {
		if _, err := fs.Stat(fsys, joinEmbed(root, "index.html")); err == nil {
			return root
		}
		if _, err := fs.Stat(fsys, joinEmbed(root, "session-page.html")); err == nil {
			return root
		}
	}
	return "."
}

// --- HTTP ---

func runHTTP(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.HTTPProbe == "" {
		t.Fatal("HTTPProbe must be set")
	}
	srv, cleanup, err := startAgentServer(t, req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{
		RealSessionID: srv.SessionID,
		BaseURL:       srv.BaseURL,
	}

	switch req.HTTPProbe {
	case HTTPProbeGoBoot, HTTPProbeInstall:
		u := srv.BaseURL + "/go"
		if srv.SessionID != "" {
			u = u + "?session=" + url.QueryEscape(srv.SessionID)
		}
		return fillGET(resp, u)
	case HTTPProbeRoot:
		u := srv.BaseURL + "/"
		status, ct, body, loc, err := doGETNoFollow(u)
		if err != nil {
			return resp, err
		}
		resp.StatusCode = status
		resp.ContentType = ct
		resp.Body = body
		resp.BodyString = string(body)
		resp.ProbeURL = u
		resp.RedirectURL = loc
		// If redirect with empty body, fetch target for product asserts.
		if len(body) == 0 && loc != "" && status >= 300 && status < 400 {
			target := loc
			if strings.HasPrefix(loc, "/") {
				target = srv.BaseURL + loc
			}
			return fillGET(resp, target)
		}
		// Same handler as /go often returns 200 HTML directly (current behavior).
		return resp, nil
	case HTTPProbeAsset:
		assetPath := req.AssetPath
		if assetPath == "" {
			assetPath = "/assets/session-page.js"
		}
		if !strings.HasPrefix(assetPath, "/") {
			assetPath = "/" + assetPath
		}
		return fillGET(resp, srv.BaseURL+assetPath)
	default:
		return nil, fmt.Errorf("unknown HTTPProbe %q", req.HTTPProbe)
	}
}

func fillGET(resp *Response, rawURL string) (*Response, error) {
	status, ct, body, err := doGET(rawURL)
	if err != nil {
		return resp, err
	}
	resp.StatusCode = status
	resp.ContentType = ct
	resp.Body = body
	resp.BodyString = string(body)
	resp.ProbeURL = rawURL
	return resp, nil
}

// --- Skill CLI ---

func runSkill(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	args := req.CLIArgs
	if len(args) == 0 {
		switch req.SkillAction {
		case SkillActionList:
			args = []string{"skill", "--list"}
		case SkillActionShow:
			args = []string{"skill", "--show"}
		case SkillActionBare:
			args = []string{"skill"}
		default:
			t.Fatalf("SkillAction %q and CLIArgs empty", req.SkillAction)
		}
	}
	req2 := *req
	req2.CLIArgs = args
	return invokeHandleCLI(t, &req2)
}

func invokeHandleCLI(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	maxWait := req.MaxDispatchWait
	if maxWait <= 0 {
		maxWait = 3 * time.Second
	}
	var stdout, stderr bytes.Buffer
	env := req.CLIEnv
	if env == nil {
		env = map[string]string{}
	}
	args := req.CLIArgs
	if args == nil {
		args = []string{}
	}

	type outcome struct{ err error }
	done := make(chan outcome, 1)
	go func() {
		done <- outcome{err: browseragent.HandleCLI(args, env, &stdout, &stderr)}
	}()

	resp := &Response{}
	select {
	case out := <-done:
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		if out.err != nil {
			resp.CLIErr = out.err.Error()
			resp.ErrText = out.err.Error()
			resp.ExitCode = 1
			return resp, nil
		}
		resp.ExitCode = 0
		return resp, nil
	case <-time.After(maxWait):
		resp.DispatchTimedOut = true
		resp.Stdout = stdout.String()
		resp.Stderr = stderr.String()
		resp.ExitCode = 1
		resp.ErrText = "HandleCLI timed out (possible accidental serve hang)"
		return resp, fmt.Errorf("%s", resp.ErrText)
	}
}

// --- Boot JSON pure ---

func runBootJSON(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	sid := req.BootSessionID
	if sid == "" {
		sid = req.SessionID
	}
	if sid == "" {
		sid = "boot-sess-test"
	}
	raw := browseragent.FormatSessionBootJSON(sid)
	resp := &Response{
		BootJSON:      raw,
		BodyString:    raw,
		BootSessionID: sid,
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		resp.BootParseOK = false
		resp.ErrText = err.Error()
		return resp, nil
	}
	resp.BootParseOK = true
	if v, ok := m["session_id"].(string); ok {
		resp.BootSessionID = v
	}
	if v, ok := m["product"].(string); ok {
		resp.BootProduct = v
	}
	switch v := m["control_port"].(type) {
	case float64:
		resp.BootControlPort = int(v)
		resp.BootPortStr = fmt.Sprintf("%d", int(v))
	case string:
		resp.BootPortStr = v
		fmt.Sscanf(v, "%d", &resp.BootControlPort)
	}
	return resp, nil
}

// --- server harness ---

type agentServer struct {
	BaseURL   string
	SessionID string
	cancel    context.CancelFunc
	done      <-chan error
}

func startAgentServer(t *testing.T, req *Request) (*agentServer, func(), error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by Setup")
	}
	if req.SessionID == "" {
		t.Fatal("SessionID must be set by Setup")
	}
	addr := req.Addr
	if addr == "" {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, nil, err
		}
		addr = ln.Addr().String()
		_ = ln.Close()
		req.Addr = addr
	}
	ready := req.ReadyTimeout
	if ready <= 0 {
		ready = 5 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())
	var stdout, stderr bytes.Buffer
	cfg := browseragent.Config{
		Addr:         addr,
		BaseDir:      req.BaseDir,
		SessionID:    req.SessionID,
		NoOpenChrome: true,
		NoAgentRun:   true,
		Stdout:       &stdout,
		Stderr:       &stderr,
	}

	done := make(chan error, 1)
	go func() {
		_, err := browseragent.Run(ctx, cfg)
		done <- err
	}()

	baseURL := "http://" + addr
	if err := waitHealth(baseURL, ready); err != nil {
		cancel()
		<-done
		return nil, nil, fmt.Errorf("control server never healthy at %s: %w", baseURL, err)
	}

	srv := &agentServer{
		BaseURL:   baseURL,
		SessionID: req.SessionID,
		cancel:    cancel,
		done:      done,
	}
	cleanup := func() {
		cancel()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
		}
	}
	return srv, cleanup, nil
}

func waitHealth(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var last error
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v1/health", nil)
		if err != nil {
			cancel()
			return err
		}
		res, err := http.DefaultClient.Do(req)
		if err == nil {
			io.Copy(io.Discard, res.Body)
			res.Body.Close()
			cancel()
			if res.StatusCode == http.StatusOK {
				return nil
			}
			last = fmt.Errorf("health status %d", res.StatusCode)
		} else {
			last = err
			cancel()
		}
		time.Sleep(20 * time.Millisecond)
	}
	if last == nil {
		last = fmt.Errorf("timeout waiting for health")
	}
	return last
}

func doGET(rawURL string) (int, string, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, "", nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, "", nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, res.Header.Get("Content-Type"), nil, err
	}
	return res.StatusCode, res.Header.Get("Content-Type"), body, nil
}

func doGETNoFollow(rawURL string) (status int, contentType string, body []byte, location string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, "", nil, "", err
	}
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 3 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return 0, "", nil, "", err
	}
	defer res.Body.Close()
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, res.Header.Get("Content-Type"), nil, "", err
	}
	return res.StatusCode, res.Header.Get("Content-Type"), body, res.Header.Get("Location"), nil
}
```
