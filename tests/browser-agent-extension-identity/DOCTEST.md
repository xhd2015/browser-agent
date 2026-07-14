# browser-agent extension identity — bundle-sum.js + serve log + hello match

Exercises **extension identity** surfaces on package
`github.com/xhd2015/browser-agent/browseragent` (after serve-runtime
/ product-hardening):

| Surface | What is under test |
|---------|-------------------|
| ParseBundleSumJS | Parse SW-loadable `bundle-sum.js` → `BundleSum{Version,MD5}` |
| ComputeExtensionContentMD5 | Sorted-path content hash of extension dir; **skip** `bundle-sum.js` |
| WriteBundleSumJS | Write `dir/bundle-sum.js` with version + md5; round-trip parse |
| Serve identity | After extract: store embedded version/md5; log; meta + `GET /v1/session` |
| Hello match | Fake WS hello with version + optional `bundle_md5` → `extension_match` |
| ColorOrangeIfTTY | Orange ANSI (`38;5;208`) when isTTY; plain otherwise |

**No production code in this tree.** **No real Chrome.** Pure helpers + in-process
`browseragent.Run` + fake WebSocket hello only.

Related regressions (run after implement, not part of this tree):

```sh
doctest test ./tests/browser-agent/...
doctest test ./tests/browser-agent-serve-runtime/...
doctest test ./tests/browser-agent-product-hardening/...
```

## Version

0.0.2

# DSN (Domain Specific Notion)

**Operator** runs **`browser-agent serve`**. After extract, the **Serve Runtime**
must know **which extension package** was staged, and later whether the
**Extension Agent** that hellos matches that package.

### 1. Bundle sum file (`bundle-sum.js`)

Generated JS (service-worker loadable via `importScripts`), Go-parseable
without a JS engine:

```text
// browser-agent bundle-sum — generated; do not edit
var BROWSER_AGENT_BUNDLE_VERSION = "1.0.1";
var BROWSER_AGENT_BUNDLE_MD5 = "a1b2c3d4e5f6789012345678abcdef01";
```

```text
BundleSum { Version string; MD5 string }  // MD5: lowercase hex, 32 chars

ParseBundleSumJS(data []byte) (BundleSum, error)
ReadBundleSumFromDir(dir string) (BundleSum, error)  // dir/bundle-sum.js
WriteBundleSumJS(dir, version, md5 string) error
```

### 2. Content MD5 (build / stage)

```text
ComputeExtensionContentMD5(dir string) (string, error)
  # sorted relative paths; skip bundle-sum.js
  # md5 over path\n + file contents + \n (or documented equivalent)
```

After staging extension files: compute md5 **excluding** sum file, then
`WriteBundleSumJS` with version from manifest + that md5.

### 3. Serve embeds identity

On `Run` after extract:

1. Ensure sum exists (write if missing from extract tree)
2. Store on session: embedded version + md5
3. Log to stderr (shape flexible if tokens present):

```text
browser-agent:   extension   <path>
browser-agent:     embedded  version=<v>  md5=<md5>
```

4. `meta.json` includes `extension_version`, `extension_md5`
5. `GET /v1/session` includes:

```json
{
  "bundled_extension": { "version": "...", "md5": "...", "path": "..." },
  "extension": {
    "connected": false,
    "version": "",
    "features": [],
    "bundle_md5": "",
    "supports_browser_agent": false
  },
  "extension_match": "not_connected|ok|version_mismatch|md5_mismatch|md5_unknown"
}
```

### 4. Hello carries loaded identity

WS hello payload:

```json
{ "version": "1.0.1", "features": ["browser-agent"], "bundle_md5": "..." }
```

Server compares loaded vs embedded → `extension_match`. On version/md5
mismatch or md5_unknown: **warning** on stderr (does **not** fail jobs in v1).
Match=ok: connected log without orange mismatch warning.

### 5. Orange TTY coloring

```text
FormatExtensionMismatchWarning(embedded, loaded BundleSum, installPath) string
ColorOrangeIfTTY(s string, isTTY bool) string
  # when isTTY: wrap with \x1b[38;5;208m … \x1b[0m
  # when !isTTY: return s unchanged
```

**Test Client** in this tree:

- Pure leaves call `ParseBundleSumJS`, `ComputeExtensionContentMD5`,
  `WriteBundleSumJS`, `ColorOrangeIfTTY` on temp dirs / fixture bytes.
- Session-match leaves start `browseragent.Run` (NoOpenChrome, NoAgentRun),
  capture stderr, optionally dial fake WS hello with version/`bundle_md5`,
  then `GET /v1/session` and parse `bundled_extension` + `extension_match`.

## Decision Tree

```
browser-agent-extension-identity
├── parse-bundle-sum/                          [ParseBundleSumJS]
│   ├── valid/                                   A1 version + md5
│   └── invalid/                                 A2 empty/malformed → error
├── compute-content-md5/                       [ComputeExtensionContentMD5]
│   ├── stable-twice/                            B1 same dir twice → equal
│   ├── changes-on-file-edit/                    B2 mutate file → md5 changes
│   └── ignores-bundle-sum/                      B+ writing sum file does not change md5
├── write-bundle-sum/                          [WriteBundleSumJS]
│   └── round-trip-parse/                        B3 write then Parse → same
├── session-match/                             [serve + GET /v1/session + fake hello]
│   ├── no-hello-not-connected/                  C1 bundled_extension set; match=not_connected
│   ├── hello-match-ok/                          C2 match=ok
│   ├── hello-version-mismatch/                  C3 match=version_mismatch; warning text
│   ├── hello-md5-mismatch/                      C4 match=md5_mismatch
│   └── hello-md5-unknown/                       C5 hello without md5 → md5_unknown
└── color-orange/                              [ColorOrangeIfTTY]
    ├── when-tty/                                D1 contains ESC + 208
    └── when-not-tty/                            D2 no ESC
```

### Parameter significance (high → low)

1. **Surface / Mode** — parse vs compute-md5 vs write-sum vs session-match vs
   color (different `Run` contracts; one Mode per top branch).
2. **Within pure hash/parse** — valid vs invalid; stable vs mutate vs exclude-sum.
3. **Within session-match** — no hello vs hello outcome
   (`not_connected|ok|version_mismatch|md5_mismatch|md5_unknown`).
4. **Within color** — isTTY true vs false.
5. **Leaf details** — fixture strings, settle timing, warning substring tokens.

## Test Index

| Leaf | Scenario |
|------|----------|
| `parse-bundle-sum/valid` | (A1) valid `bundle-sum.js` bytes → Version + MD5 |
| `parse-bundle-sum/invalid` | (A2) empty / missing tokens / garbage → non-nil error |
| `compute-content-md5/stable-twice` | (B1) same fixture dir hashed twice → equal 32-hex md5 |
| `compute-content-md5/changes-on-file-edit` | (B2) change one file content → md5 changes |
| `compute-content-md5/ignores-bundle-sum` | (B+) hash, write `bundle-sum.js`, re-hash → same md5 |
| `write-bundle-sum/round-trip-parse` | (B3) WriteBundleSumJS then ParseBundleSumJS → same version+md5 |
| `session-match/no-hello-not-connected` | (C1) serve, no WS → `bundled_extension` set; `extension_match=not_connected` |
| `session-match/hello-match-ok` | (C2) hello version+md5 match embedded → `extension_match=ok` |
| `session-match/hello-version-mismatch` | (C3) hello wrong version → `version_mismatch`; stderr warning has embedded+loaded |
| `session-match/hello-md5-mismatch` | (C4) hello wrong md5 → `md5_mismatch` |
| `session-match/hello-md5-unknown` | (C5) hello without `bundle_md5` → `md5_unknown` |
| `color-orange/when-tty` | (D1) ColorOrangeIfTTY(s,true) wraps with ESC + `38;5;208` |
| `color-orange/when-not-tty` | (D2) ColorOrangeIfTTY(s,false) equals plain s (no ESC) |

**Leaf count: 13**

## How to Run

```sh
cd project-api-capture
doctest vet ./tests/browser-agent-extension-identity
doctest test ./tests/browser-agent-extension-identity/...
# or:
cd tests/browser-agent-extension-identity && doctest vet . && doctest test -v .
```

### Implementer contract (authoritative for GREEN)

```text
type BundleSum struct {
    Version string
    MD5     string // lowercase hex, 32 chars preferred
}

func ParseBundleSumJS(data []byte) (BundleSum, error)
func ReadBundleSumFromDir(dir string) (BundleSum, error) // reads dir/bundle-sum.js
func WriteBundleSumJS(dir string, version, md5 string) error
func ComputeExtensionContentMD5(dir string) (string, error)
// skip basename bundle-sum.js; sorted relative paths; deterministic digest

func FormatExtensionMismatchWarning(embedded, loaded BundleSum, installPath string) string
func ColorOrangeIfTTY(s string, isTTY bool) string
// orange: \x1b[38;5;208m … \x1b[0m when isTTY; plain otherwise

// Serve after extract:
// - ensure bundle-sum.js (write if missing)
// - session stores embedded version+md5
// - stderr logs embedded version + md5
// - meta.json: extension_version, extension_md5 (+ existing fields)
// - GET /v1/session: bundled_extension{version,md5,path}, extension.bundle_md5,
//   extension_match ∈ {not_connected,ok,version_mismatch,md5_mismatch,md5_unknown}
//
// WS hello payload may include bundle_md5; markHello accepts md5 and sets match.
// Mismatch → warning on stderr (orange when stderr TTY); do NOT fail jobs in v1.
```

```go
import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode values — top-level surface under test.
const (
	ModeParseBundleSum   = "parse-bundle-sum"
	ModeComputeContentMD5 = "compute-content-md5"
	ModeWriteBundleSum   = "write-bundle-sum"
	ModeSessionMatch     = "session-match"
	ModeColorOrange      = "color-orange"
)

// ComputeProbe for ModeComputeContentMD5.
const (
	ComputeProbeStableTwice      = "stable-twice"
	ComputeProbeChangesOnEdit    = "changes-on-file-edit"
	ComputeProbeIgnoresBundleSum = "ignores-bundle-sum"
)

// SessionMatchKind for ModeSessionMatch.
const (
	SessionMatchNoHello           = "no-hello-not-connected"
	SessionMatchHelloOK           = "hello-match-ok"
	SessionMatchVersionMismatch   = "hello-version-mismatch"
	SessionMatchMD5Mismatch       = "hello-md5-mismatch"
	SessionMatchMD5Unknown        = "hello-md5-unknown"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	// Mode selects the surface Run executes.
	Mode string

	// ModuleRoot is project-api-capture module directory.
	// Root Setup sets from DOCTEST_ROOT/../..
	ModuleRoot string

	// --- parse-bundle-sum ---
	BundleSumJS []byte

	// --- compute / write (temp extension dir) ---
	// ExtensionDir is filled by Run harness when empty (temp fixture dir).
	ExtensionDir string
	// FixtureFiles maps relative path → content for staged extension dir.
	// Default fixture applied when empty.
	FixtureFiles map[string]string
	// ComputeProbe: stable-twice | changes-on-file-edit | ignores-bundle-sum
	ComputeProbe string
	// EditRelPath / EditNewContent used by changes-on-file-edit.
	EditRelPath   string
	EditNewContent string

	// WriteVersion / WriteMD5 for WriteBundleSumJS leaves.
	WriteVersion string
	WriteMD5     string

	// --- session-match ---
	BaseDir      string
	Addr         string
	SessionID    string
	NoOpenChrome bool
	NoAgentRun   bool
	ReadyTimeout time.Duration
	HelloSettle  time.Duration

	// SessionMatchKind selects C1–C5 outcome.
	SessionMatchKind string

	// Fake hello knobs (session-match).
	DoHello         bool
	HelloVersion    string
	HelloMD5        string
	HelloOmitMD5    bool // when true, hello payload has no bundle_md5
	HelloFeatures   []string
	// ForceHelloMD5 / ForceHelloVersion override embedded identity for mismatch cases.
	// When empty, Run uses embedded values from session for match-ok.
	ForceHelloVersion string
	ForceHelloMD5     string

	// --- color-orange ---
	ColorInput string
	IsTTY      bool
}

// Response holds outcomes for all modes (fields used per mode).
type Response struct {
	// Parse / write
	ParseOK      bool
	ParseErr     string
	Version      string
	MD5          string
	BundleSumPath string
	WrittenJS    string

	// Compute
	MD5First  string
	MD5Second string
	MD5Equal  bool
	ExtensionDir string

	// Color
	ColorOut string

	// Serve / session
	BaseURL           string
	RealSessionID     string
	Addr              string
	StatusCode        int
	ContentType       string
	Body              []byte
	BodyString        string
	Raw               map[string]any
	Meta              map[string]any
	MetaJSON          string
	MetaPath          string
	SessionDir        string
	Stderr            string
	Stdout            string
	ExtensionInstallPath string

	// Snapshot identity fields
	BundledVersion string
	BundledMD5     string
	BundledPath    string
	ExtensionMatch string
	ExtConnected   bool
	ExtVersion     string
	ExtBundleMD5   string
	ExtSupports    bool
	MetaExtVersion string
	MetaExtMD5     string

	// Warning helpers
	WarningText string
	HasWarning  bool

	ErrText  string
	ExitCode int
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
	case ModeParseBundleSum:
		return runParseBundleSum(t, req)
	case ModeComputeContentMD5:
		return runComputeContentMD5(t, req)
	case ModeWriteBundleSum:
		return runWriteBundleSum(t, req)
	case ModeSessionMatch:
		return runSessionMatch(t, req)
	case ModeColorOrange:
		return runColorOrange(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

// --- parse ---

func runParseBundleSum(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	data := req.BundleSumJS
	sum, err := browseragent.ParseBundleSumJS(data)
	resp := &Response{
		Version: sum.Version,
		MD5:     sum.MD5,
	}
	if err != nil {
		resp.ParseOK = false
		resp.ParseErr = err.Error()
		resp.ErrText = err.Error()
		resp.ExitCode = 1
		return resp, nil
	}
	resp.ParseOK = true
	resp.ExitCode = 0
	return resp, nil
}

// --- compute md5 ---

func runComputeContentMD5(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	probe := req.ComputeProbe
	if probe == "" {
		t.Fatal("ComputeProbe must be set")
	}
	dir, err := ensureFixtureExtDir(t, req)
	if err != nil {
		return nil, err
	}
	resp := &Response{ExtensionDir: dir}

	switch probe {
	case ComputeProbeStableTwice:
		h1, err := browseragent.ComputeExtensionContentMD5(dir)
		if err != nil {
			resp.ErrText = err.Error()
			resp.ExitCode = 1
			return resp, nil
		}
		h2, err := browseragent.ComputeExtensionContentMD5(dir)
		if err != nil {
			resp.ErrText = err.Error()
			resp.ExitCode = 1
			return resp, nil
		}
		resp.MD5First = h1
		resp.MD5Second = h2
		resp.MD5Equal = h1 == h2
		resp.MD5 = h1
		resp.ExitCode = 0
		return resp, nil

	case ComputeProbeChangesOnEdit:
		h1, err := browseragent.ComputeExtensionContentMD5(dir)
		if err != nil {
			resp.ErrText = err.Error()
			resp.ExitCode = 1
			return resp, nil
		}
		rel := req.EditRelPath
		if rel == "" {
			rel = "contentScript.js"
		}
		newContent := req.EditNewContent
		if newContent == "" {
			newContent = "// mutated content for md5 change\nconsole.log('mutated');\n"
		}
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.WriteFile(path, []byte(newContent), 0o644); err != nil {
			return resp, fmt.Errorf("mutate %s: %w", path, err)
		}
		h2, err := browseragent.ComputeExtensionContentMD5(dir)
		if err != nil {
			resp.ErrText = err.Error()
			resp.ExitCode = 1
			return resp, nil
		}
		resp.MD5First = h1
		resp.MD5Second = h2
		resp.MD5Equal = h1 == h2
		resp.MD5 = h1
		resp.ExitCode = 0
		return resp, nil

	case ComputeProbeIgnoresBundleSum:
		h1, err := browseragent.ComputeExtensionContentMD5(dir)
		if err != nil {
			resp.ErrText = err.Error()
			resp.ExitCode = 1
			return resp, nil
		}
		// Write bundle-sum.js (must be excluded from content hash).
		ver := req.WriteVersion
		if ver == "" {
			ver = "1.0.1"
		}
		sumMD5 := req.WriteMD5
		if sumMD5 == "" {
			sumMD5 = h1
		}
		if err := browseragent.WriteBundleSumJS(dir, ver, sumMD5); err != nil {
			// If Write not implemented yet, fall back to raw write so the exclude
			// contract of Compute can still be asserted once Write lands.
			js := formatBundleSumJS(ver, sumMD5)
			if werr := os.WriteFile(filepath.Join(dir, "bundle-sum.js"), []byte(js), 0o644); werr != nil {
				return resp, fmt.Errorf("write bundle-sum.js: WriteBundleSumJS=%v raw=%w", err, werr)
			}
			resp.ParseErr = err.Error() // note Write failure for diagnostics
		}
		h2, err := browseragent.ComputeExtensionContentMD5(dir)
		if err != nil {
			resp.ErrText = err.Error()
			resp.ExitCode = 1
			return resp, nil
		}
		resp.MD5First = h1
		resp.MD5Second = h2
		resp.MD5Equal = h1 == h2
		resp.MD5 = h1
		resp.BundleSumPath = filepath.Join(dir, "bundle-sum.js")
		resp.ExitCode = 0
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown ComputeProbe %q", probe)
	}
}

// --- write + parse round-trip ---

func runWriteBundleSum(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	dir, err := ensureFixtureExtDir(t, req)
	if err != nil {
		return nil, err
	}
	ver := req.WriteVersion
	if ver == "" {
		ver = "1.0.1"
	}
	md5hex := req.WriteMD5
	if md5hex == "" {
		md5hex = "a1b2c3d4e5f6789012345678abcdef01"
	}
	resp := &Response{
		ExtensionDir:  dir,
		BundleSumPath: filepath.Join(dir, "bundle-sum.js"),
		Version:       ver,
		MD5:           md5hex,
	}
	if err := browseragent.WriteBundleSumJS(dir, ver, md5hex); err != nil {
		resp.ParseOK = false
		resp.ParseErr = err.Error()
		resp.ErrText = err.Error()
		resp.ExitCode = 1
		return resp, nil
	}
	data, rerr := os.ReadFile(resp.BundleSumPath)
	if rerr != nil {
		resp.ParseOK = false
		resp.ParseErr = rerr.Error()
		resp.ErrText = rerr.Error()
		resp.ExitCode = 1
		return resp, nil
	}
	resp.WrittenJS = string(data)
	sum, perr := browseragent.ParseBundleSumJS(data)
	if perr != nil {
		resp.ParseOK = false
		resp.ParseErr = perr.Error()
		resp.ErrText = perr.Error()
		resp.ExitCode = 1
		return resp, nil
	}
	resp.ParseOK = true
	resp.Version = sum.Version
	resp.MD5 = sum.MD5
	resp.ExitCode = 0
	return resp, nil
}

// --- session match (serve + optional hello) ---

func runSessionMatch(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	kind := req.SessionMatchKind
	if kind == "" {
		t.Fatal("SessionMatchKind must be set")
	}
	req.NoOpenChrome = true
	req.NoAgentRun = true

	// Capture stderr for embedded log + mismatch warnings.
	var stdout, stderr bytes.Buffer
	srv, cleanup, err := startAgentServer(t, req, &stdout, &stderr)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp := &Response{
		BaseURL:       srv.BaseURL,
		RealSessionID: srv.SessionID,
		Addr:          srv.Addr,
		SessionDir:    filepath.Join(req.BaseDir, "sessions", srv.SessionID),
		Stdout:        stdout.String(),
		Stderr:        stderr.String(),
	}

	// Load meta.json early (extension identity fields).
	metaPath := filepath.Join(resp.SessionDir, "meta.json")
	resp.MetaPath = metaPath
	if data, err := os.ReadFile(metaPath); err == nil {
		resp.MetaJSON = string(data)
		var m map[string]any
		if json.Unmarshal(data, &m) == nil {
			resp.Meta = m
			resp.MetaExtVersion = stringField(m, "extension_version", "extensionVersion")
			resp.MetaExtMD5 = stringField(m, "extension_md5", "extensionMd5", "extensionMD5")
			resp.ExtensionInstallPath = stringField(m, "extension_install_path", "extensionInstallPath")
		}
	}

	// Probe session before hello for not_connected; after hello for others.
	// First read snapshot to learn embedded identity (for match-ok construction).
	if err := getAndFillSession(resp, srv); err != nil {
		return resp, err
	}

	// Refresh stderr after settle (serve may log embedded on start).
	resp.Stderr = stderr.String()
	resp.Stdout = stdout.String()

	needHello := kind != SessionMatchNoHello
	if needHello || req.DoHello {
		version, md5hex, omitMD5 := helloPayloadForKind(kind, req, resp)
		feats := req.HelloFeatures
		if len(feats) == 0 {
			feats = []string{"browser-agent"}
		}
		ext, derr := dialFakeExtensionHello(srv.BaseURL, version, md5hex, omitMD5, feats)
		if derr != nil {
			return resp, fmt.Errorf("fake extension dial: %w", derr)
		}
		defer ext.Close()
		// Keep conn open so session stays connected while we probe.
		go ext.Loop()
		settle := req.HelloSettle
		if settle <= 0 {
			settle = 80 * time.Millisecond
		}
		time.Sleep(settle)
		// Re-probe session after hello.
		if err := getAndFillSession(resp, srv); err != nil {
			return resp, err
		}
		resp.Stderr = stderr.String()
		resp.Stdout = stdout.String()
		resp.WarningText = extractMismatchWarning(resp.Stderr)
		resp.HasWarning = resp.WarningText != "" ||
			strings.Contains(strings.ToLower(resp.Stderr), "mismatch") ||
			strings.Contains(strings.ToLower(resp.Stderr), "md5_unknown") ||
			strings.Contains(strings.ToLower(resp.Stderr), "version_mismatch") ||
			strings.Contains(strings.ToLower(resp.Stderr), "md5_mismatch")
	}

	resp.ExitCode = 0
	return resp, nil
}

func helloPayloadForKind(kind string, req *Request, resp *Response) (version, md5hex string, omitMD5 bool) {
	// Prefer explicit leaf overrides.
	version = req.ForceHelloVersion
	if version == "" {
		version = req.HelloVersion
	}
	md5hex = req.ForceHelloMD5
	if md5hex == "" {
		md5hex = req.HelloMD5
	}
	omitMD5 = req.HelloOmitMD5

	// Defaults from embedded identity when available.
	embeddedV := resp.BundledVersion
	if embeddedV == "" {
		embeddedV = resp.MetaExtVersion
	}
	if embeddedV == "" {
		embeddedV = "1.0.1"
	}
	embeddedMD5 := resp.BundledMD5
	if embeddedMD5 == "" {
		embeddedMD5 = resp.MetaExtMD5
	}

	switch kind {
	case SessionMatchHelloOK:
		if version == "" {
			version = embeddedV
		}
		if md5hex == "" {
			md5hex = embeddedMD5
		}
		omitMD5 = false
	case SessionMatchVersionMismatch:
		if version == "" {
			// Distinct from embedded.
			version = "9.9.9"
			if version == embeddedV {
				version = "0.0.1"
			}
		}
		if md5hex == "" {
			md5hex = embeddedMD5
		}
		omitMD5 = false
	case SessionMatchMD5Mismatch:
		if version == "" {
			version = embeddedV
		}
		if md5hex == "" {
			// 32-hex distinct from embedded when known.
			md5hex = "ffffffffffffffffffffffffffffffff"
			if md5hex == embeddedMD5 {
				md5hex = "00000000000000000000000000000000"
			}
		}
		omitMD5 = false
	case SessionMatchMD5Unknown:
		if version == "" {
			version = embeddedV
		}
		md5hex = ""
		omitMD5 = true
	default:
		if version == "" {
			version = embeddedV
		}
	}
	return version, md5hex, omitMD5
}

func getAndFillSession(resp *Response, srv *agentServer) error {
	u := srv.BaseURL + "/v1/session?session=" + url.QueryEscape(srv.SessionID)
	status, ct, body, err := doGET(u)
	resp.StatusCode = status
	resp.ContentType = ct
	resp.Body = body
	resp.BodyString = string(body)
	if err != nil {
		return err
	}
	fillSessionIdentity(resp, body)
	return nil
}

func fillSessionIdentity(resp *Response, body []byte) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return
	}
	resp.Raw = raw
	resp.ExtensionMatch = stringField(raw, "extension_match", "extensionMatch")
	if be, ok := raw["bundled_extension"].(map[string]any); ok {
		resp.BundledVersion = stringField(be, "version")
		resp.BundledMD5 = stringField(be, "md5", "MD5")
		resp.BundledPath = stringField(be, "path")
		if resp.ExtensionInstallPath == "" {
			resp.ExtensionInstallPath = resp.BundledPath
		}
	}
	// Also accept top-level aliases if implementer flattens.
	if resp.BundledVersion == "" {
		resp.BundledVersion = stringField(raw, "extension_version", "embedded_version")
	}
	if resp.BundledMD5 == "" {
		resp.BundledMD5 = stringField(raw, "extension_md5", "embedded_md5")
	}
	if ext, ok := raw["extension"].(map[string]any); ok {
		if c, ok := ext["connected"].(bool); ok {
			resp.ExtConnected = c
		}
		resp.ExtVersion = stringField(ext, "version")
		resp.ExtBundleMD5 = stringField(ext, "bundle_md5", "bundleMd5", "bundleMD5", "md5")
		if s, ok := ext["supports_browser_agent"].(bool); ok {
			resp.ExtSupports = s
		}
	}
}

// --- color ---

func runColorOrange(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	in := req.ColorInput
	if in == "" {
		in = "browser-agent: warning: extension identity mismatch"
	}
	out := browseragent.ColorOrangeIfTTY(in, req.IsTTY)
	return &Response{
		ColorOut: out,
		ExitCode: 0,
	}, nil
}

// --- fixture helpers ---

func defaultFixtureFiles() map[string]string {
	return map[string]string{
		"manifest.json": `{
  "manifest_version": 3,
  "name": "Browser Agent Fixture",
  "version": "1.0.1",
  "background": { "service_worker": "background.js" },
  "permissions": ["alarms", "debugger", "storage", "tabs"],
  "host_permissions": ["http://127.0.0.1:43761/*", "<all_urls>"]
}
`,
		"background.js":    "// fixture background\nconsole.log('bg');\n",
		"contentScript.js": "// fixture content\nconsole.log('cs');\n",
		"popup.html":       "<!doctype html><title>fixture</title>\n",
	}
}

func ensureFixtureExtDir(t *testing.T, req *Request) (string, error) {
	t.Helper()
	if req.ExtensionDir != "" {
		if err := os.MkdirAll(req.ExtensionDir, 0o755); err != nil {
			return "", err
		}
		// If already populated (has any file), reuse; else write fixtures.
		entries, _ := os.ReadDir(req.ExtensionDir)
		if len(entries) > 0 {
			return req.ExtensionDir, nil
		}
	} else {
		dir := filepath.Join(t.TempDir(), "ext")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
		req.ExtensionDir = dir
	}
	files := req.FixtureFiles
	if len(files) == 0 {
		files = defaultFixtureFiles()
	}
	for rel, content := range files {
		path := filepath.Join(req.ExtensionDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return "", err
		}
	}
	return req.ExtensionDir, nil
}

func formatBundleSumJS(version, md5hex string) string {
	return fmt.Sprintf(
		"// browser-agent bundle-sum — generated; do not edit\nvar BROWSER_AGENT_BUNDLE_VERSION = %q;\nvar BROWSER_AGENT_BUNDLE_MD5 = %q;\n",
		version, md5hex,
	)
}

func extractMismatchWarning(stderr string) string {
	if stderr == "" {
		return ""
	}
	var b strings.Builder
	for _, line := range strings.Split(stderr, "\n") {
		low := strings.ToLower(line)
		if strings.Contains(low, "mismatch") ||
			strings.Contains(low, "md5_unknown") ||
			strings.Contains(low, "identity") && strings.Contains(low, "warn") ||
			strings.Contains(low, "bundle") && (strings.Contains(low, "warn") || strings.Contains(low, "mismatch")) {
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	return strings.TrimSpace(b.String())
}

// --- serve + fake extension ---

type agentServer struct {
	BaseURL   string
	SessionID string
	BaseDir   string
	Addr      string
	cancel    context.CancelFunc
}

func startAgentServer(t *testing.T, req *Request, stdout, stderr io.Writer) (*agentServer, func(), error) {
	t.Helper()
	baseDir := req.BaseDir
	if baseDir == "" {
		baseDir = filepath.Join(t.TempDir(), "ba-base")
		if err := os.MkdirAll(baseDir, 0o755); err != nil {
			return nil, nil, err
		}
		req.BaseDir = baseDir
	}
	sid := req.SessionID
	if sid == "" {
		sid = fmt.Sprintf("sess-extid-%d", time.Now().UnixNano()%1e12)
		req.SessionID = sid
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
	readyTO := req.ReadyTimeout
	if readyTO <= 0 {
		readyTO = 5 * time.Second
	}
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	ctx, cancel := context.WithCancel(context.Background())
	cfg := browseragent.Config{
		Addr:         addr,
		BaseDir:      baseDir,
		SessionID:    sid,
		NoOpenChrome: true,
		NoAgentRun:   true,
		ReadyTimeout: readyTO,
		Stdout:       stdout,
		Stderr:       stderr,
	}
	errCh := make(chan error, 1)
	go func() {
		_, err := browseragent.Run(ctx, cfg)
		errCh <- err
	}()
	baseURL := "http://" + addr
	if err := waitHealth(baseURL, readyTO); err != nil {
		cancel()
		select {
		case <-errCh:
		case <-time.After(2 * time.Second):
		}
		return nil, nil, fmt.Errorf("serve health: %w", err)
	}
	// Brief settle so extract + identity log + meta write finish.
	time.Sleep(80 * time.Millisecond)
	srv := &agentServer{
		BaseURL:   baseURL,
		SessionID: sid,
		BaseDir:   baseDir,
		Addr:      addr,
		cancel:    cancel,
	}
	cleanup := func() {
		cancel()
		select {
		case <-errCh:
		case <-time.After(3 * time.Second):
		}
	}
	return srv, cleanup, nil
}

func waitHealth(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var last error
	for time.Now().Before(deadline) {
		resp, err := http.Get(strings.TrimRight(baseURL, "/") + "/v1/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == 200 {
				return nil
			}
			last = fmt.Errorf("health status %d", resp.StatusCode)
		} else {
			last = err
		}
		time.Sleep(25 * time.Millisecond)
	}
	if last == nil {
		last = fmt.Errorf("health timeout")
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

type fakeExtension struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func dialFakeExtensionHello(baseURL, version, bundleMD5 string, omitMD5 bool, features []string) (*fakeExtension, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	u.Scheme = "ws"
	u.Path = "/v1/ws"
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{
		"version":  version,
		"features": features,
	}
	if !omitMD5 {
		payload["bundle_md5"] = bundleMD5
	}
	hello := map[string]any{
		"v":       1,
		"type":    "hello",
		"payload": payload,
	}
	if err := conn.WriteJSON(hello); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return &fakeExtension{conn: conn}, nil
}

func (f *fakeExtension) Close() {
	if f != nil && f.conn != nil {
		_ = f.conn.Close()
	}
}

func (f *fakeExtension) Loop() {
	for {
		var msg map[string]any
		if err := f.conn.ReadJSON(&msg); err != nil {
			return
		}
		// Ignore server pushes for identity tests (no job completion needed).
		_ = msg
	}
}

func stringField(m map[string]any, keys ...string) string {
	if m == nil {
		return nilString()
	}
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			switch t := v.(type) {
			case string:
				if strings.TrimSpace(t) != "" {
					return t
				}
			default:
				s := fmt.Sprint(t)
				if s != "" && s != "<nil>" {
					return s
				}
			}
		}
	}
	return ""
}

func nilString() string { return "" }

// Keep imports used across harness helpers (codegen / vet friendliness).
var (
	_ = bytes.Buffer{}
	_ = context.Background
	_ = md5.Sum
	_ = hex.EncodeToString
	_ = json.Marshal
	_ = fmt.Sprintf
	_ = io.Discard
	_ = net.Listen
	_ = http.Get
	_ = url.Parse
	_ = os.TempDir
	_ = filepath.Join
	_ = strings.Contains
	_ = sync.Mutex{}
	_ = time.Second
	_ = websocket.DefaultDialer
	_ = browseragent.ProductName
)
```
