# browser-agent Firefox extension HTTP poll transport (Phase 2)

Classic TDD for **Phase 2 of 2** of the Firefox HTTP poll design: Firefox
extension **background** uses **HTTP poll** (`POST /v1/ext/hello|poll|result`) as
transport (primary or WS-failure fallback) — **not WebSocket-only**.

| Surface | What is under test |
|---------|-------------------|
| `/v1/ext/hello` | Attach without requiring a live WebSocket |
| `/v1/ext/poll` | Long-poll loop (`wait_ms`) for jobs + events |
| `/v1/ext/result` | Complete jobs over HTTP (parity with WS `type=result`) |
| HTTP client | `fetch` (or equivalent POST client) for poll transport |
| Transport mode | Not WS-only — poll path present (WS may remain) |
| Register | Session register starts poll loop / HTTP transport |
| Events | `prepare_reconnect` drained from poll events |

**No real Firefox. No network. No live control server.** Source markers on
`Firefox-Ext-Browser-Agent` background only (public preferred; build fallback).

**Classic TDD — RED** until implementer wires HTTP poll into Firefox background
(and syncs public → build → embedded/extension-firefox → fixtures).

### Out of scope

- Server routes `POST /v1/ext/*` (covered by `tests/browser-agent-ext-http-poll`)
- Chrome extension changes (Chrome stays WS)
- Removing WebSocket entirely (optional try-WS-then-fallback is OK)
- Full job handler matrix (other Firefox phase trees)

## Version

0.0.2

# DSN (Domain Specific Notion)

**Session Page** is opened in Firefox; the **Content Script** still **registers**
with the **Background Worker**. On Firefox, the background prefers **HTTP poll
transport** to the **Control Server** (plain `http://127.0.0.1:{port}`) so mixed-
content / `wss` upgrade noise is avoided. WebSocket may still exist as a try-once
or leftover path, but the session must not be **WS-only**.

```text
Content Script
  -> browser.runtime.sendMessage({ type: "register", session_id, control_port, … })

Background on register
  -> sessions[S] = { … transport, pollAbort, … }
  -> HTTP poll path (primary on Firefox; or after WS fail fallback):
       POST http://127.0.0.1:{port}/v1/ext/hello
         { session_id, version, features, browser_product: "firefox", … }
       loop:
         POST /v1/ext/poll { session_id, wait_ms }
           -> jobs[]  => handleJob => POST /v1/ext/result
           -> events[] (e.g. prepare_reconnect) => schedule close + reconnect
  -> (optional) try WebSocket first; on persistent fail / onerror switch to poll

Control Server (already GREEN server-side)
  <- hello marks extension_connected without socket
  <- poll leases jobs + drains events
  <- result completes waiters
```

**Test Client** reads `Firefox-Ext-Browser-Agent` sources and asserts tokens —
never launches a browser or binds a port.

```text
Test Client
  -> read public/background.js (or build/)
  -> assert /v1/ext/hello|poll|result, fetch/HTTP poll, not WS-only
  -> assert register starts poll transport; prepare_reconnect on poll events
```

## Decision Tree

```
browser-agent-firefox-http-poll
└── background-source/                              [Firefox-Ext public/build background.js]
    ├── hello-endpoint/                               /v1/ext/hello path token
    ├── poll-endpoint/                                /v1/ext/poll + wait_ms
    ├── result-endpoint/                              /v1/ext/result path token
    ├── http-poll-fetch/                              fetch (or POST client) for poll
    ├── transport-not-ws-only/                        anti: not WebSocket-only transport
    ├── register-starts-poll/                         register starts HTTP poll path
    └── prepare-reconnect-on-poll/                    prepare_reconnect from poll events
```

### Parameter significance (high → low)

1. **Surface / Mode** — background-source only (single file family; all leaves share probe).
2. **Transport contract aspect** — endpoint paths vs HTTP client vs anti-WS-only
   vs register integration vs poll event handling (MECE branches of the client contract).

## Test Index

| Leaf | Scenario |
|------|----------|
| `background-source/hello-endpoint` | `background.js` references **`/v1/ext/hello`** (HTTP attach) |
| `background-source/poll-endpoint` | **`/v1/ext/poll`** present; **`wait_ms`** (or `waitMs`) for long-poll |
| `background-source/result-endpoint` | **`/v1/ext/result`** present (job completion over HTTP) |
| `background-source/http-poll-fetch` | Uses **`fetch`** (or clear HTTP POST client) with poll/hello/result transport language |
| `background-source/transport-not-ws-only` | Explicit: transport is **not WS-only** — HTTP poll endpoints present (WS may remain) |
| `background-source/register-starts-poll` | **register** path starts HTTP poll / poll loop (not only `new WebSocket`) |
| `background-source/prepare-reconnect-on-poll` | **prepare_reconnect** handled in context of poll **events** (or shared reconnect path reachable from poll) |

**Leaf count: 7**

## How to Run

```sh
doctest vet ./tests/browser-agent-firefox-http-poll
doctest test ./tests/browser-agent-firefox-http-poll
# expect RED until implementer lands Firefox background HTTP poll transport
```

Module: `github.com/xhd2015/browser-agent`. Under test: `Firefox-Ext-Browser-Agent`
sources (read-only FS). Server routes are assumed available (separate tree).

### Implementer contract (authoritative for GREEN)

#### `Firefox-Ext-Browser-Agent/public/background.js`

1. **HTTP poll transport on Firefox** (recommended: primary; optional: try WS once
   then fall back on persistent failure / security upgrade / `onerror`).
   Chrome path remains WebSocket — do not require Chrome changes.

2. **Hello** — `POST` to **`/v1/ext/hello`** with at least `session_id`, plus
   version/features/`browser_product` (firefox) as today.

3. **Poll loop** — `POST` **`/v1/ext/poll`** with `session_id` and **`wait_ms`**
   (server normalizes default/cap). On response:
   - **jobs** → existing `handleJob` → **`POST /v1/ext/result`**
   - **events** → handle **`prepare_reconnect`** (and any other session events)

4. **Register** — still creates the session entry; starts the **poll loop** (or
   connect that ends in poll transport), not only `new WebSocket(...)`.

5. **WebSocket** — may remain for try-first or dual-mode; **not** required to
   remove. GREEN requires poll path present so transport is not WS-only.

6. **Sync** after public changes: `public` → `build` →
   `browseragent/embedded/extension-firefox` → fixtures when used (same practice
   as other Firefox phases). This tree probes **public first**, then build.

**Not required this tree:** live poll e2e, server handler changes, SPA install UX.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Mode — top-level surface under test.
const (
	ModeBackgroundSource = "background-source"
)

// BackgroundSourceTarget for ModeBackgroundSource.
const (
	BgSrcHelloEndpoint           = "hello-endpoint"
	BgSrcPollEndpoint            = "poll-endpoint"
	BgSrcResultEndpoint          = "result-endpoint"
	BgSrcHttpPollFetch           = "http-poll-fetch"
	BgSrcTransportNotWsOnly      = "transport-not-ws-only"
	BgSrcRegisterStartsPoll      = "register-starts-poll"
	BgSrcPrepareReconnectOnPoll  = "prepare-reconnect-on-poll"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string

	// background-source
	BackgroundSourceTarget string
}

// Response holds probe outcomes.
type Response struct {
	FoundPaths   []string
	FileExists   bool
	CombinedText string
	FileContents map[string]string
	ErrText      string
}

// Run executes the scenario selected by req.Mode and leaf Setup narrowing.
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if req == nil {
		t.Fatal("req is nil")
	}
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	if req.ModuleRoot == "" {
		req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	}
	switch req.Mode {
	case ModeBackgroundSource:
		return runBackgroundSource(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runBackgroundSource(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BackgroundSourceTarget == "" {
		t.Fatal("BackgroundSourceTarget must be set")
	}
	// All background leaves share the same file probe; Assert specializes.
	_ = req.BackgroundSourceTarget
	resp := &Response{FileContents: map[string]string{}}
	candidates := firefoxBackgroundCandidates(req.ModuleRoot)
	path, data, ok := firstExistingFile(candidates)
	resp.FileExists = ok
	if ok {
		resp.FoundPaths = []string{path}
		resp.FileContents[path] = string(data)
		resp.CombinedText = string(data)
	} else {
		resp.ErrText = "background.js not found under Firefox-Ext-Browser-Agent"
	}
	return resp, nil
}

func firefoxBackgroundCandidates(root string) []string {
	return []string{
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "public", "background.js"),
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "build", "background.js"),
		filepath.Join(root, "Firefox-Ext-Browser-Agent", "background.js"),
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

// --- source marker helpers (shared by ASSERT.md) ---

// hasExtHelloPath reports POST /v1/ext/hello transport token.
func hasExtHelloPath(src string) bool {
	return strings.Contains(src, "/v1/ext/hello")
}

// hasExtPollPath reports POST /v1/ext/poll transport token.
func hasExtPollPath(src string) bool {
	return strings.Contains(src, "/v1/ext/poll")
}

// hasExtResultPath reports POST /v1/ext/result transport token.
func hasExtResultPath(src string) bool {
	return strings.Contains(src, "/v1/ext/result")
}

// hasWaitMs reports long-poll wait parameter (snake or camel).
func hasWaitMs(src string) bool {
	return strings.Contains(src, "wait_ms") ||
		strings.Contains(src, "waitMs") ||
		strings.Contains(src, "waitMS")
}

// hasHttpPollEndpoints is true when all three HTTP poll routes appear.
func hasHttpPollEndpoints(src string) bool {
	return hasExtHelloPath(src) && hasExtPollPath(src) && hasExtResultPath(src)
}

// hasFetchClient reports use of fetch (primary WebExtensions HTTP client).
func hasFetchClient(src string) bool {
	// Match call-like forms; avoid matching only comments with the word "fetch".
	if strings.Contains(src, "fetch(") || strings.Contains(src, "fetch (") {
		return true
	}
	// globalThis.fetch / self.fetch
	if strings.Contains(src, ".fetch(") {
		return true
	}
	return false
}

// hasHttpPostMethod reports explicit POST method for HTTP poll client.
func hasHttpPostMethod(src string) bool {
	candidates := []string{
		`method: "POST"`,
		`method: 'POST'`,
		`method:"POST"`,
		`method:'POST'`,
		`"POST"`,
		`'POST'`,
		`http.MethodPost`, // unlikely in JS but harmless
	}
	for _, c := range candidates {
		if strings.Contains(src, c) {
			return true
		}
	}
	// case-insensitive method: post patterns
	low := strings.ToLower(src)
	return strings.Contains(low, `method: "post"`) ||
		strings.Contains(low, `method: 'post'`) ||
		strings.Contains(low, `"method":"post"`)
}

// hasHttpPollClientLanguage reports poll-loop / transport naming beyond bare paths.
func hasHttpPollClientLanguage(src string) bool {
	low := strings.ToLower(src)
	markers := []string{
		"httppoll",
		"http_poll",
		"http-poll",
		"pollloop",
		"poll_loop",
		"startpoll",
		"starthttppoll",
		"polltransport",
		"transportmode",
		"usehttppoll",
		"pollabort",
		"longpoll",
		"long-poll",
		"long_poll",
	}
	for _, m := range markers {
		if strings.Contains(low, m) {
			return true
		}
	}
	// path tokens themselves count as transport language when combined with fetch
	return hasExtPollPath(src) || hasExtHelloPath(src)
}

// hasHttpPollFetch reports fetch (or POST client) used for HTTP poll transport.
// Requires (fetch OR explicit POST method) AND (poll endpoints or poll language).
func hasHttpPollFetch(src string) bool {
	client := hasFetchClient(src) || hasHttpPostMethod(src)
	if !client {
		return false
	}
	return hasHttpPollClientLanguage(src) || hasExtPollPath(src) || hasExtHelloPath(src)
}

// isWsOnlyTransport is true when WebSocket control path is present but HTTP poll
// endpoints are missing — the Phase-1 connect shape that this tree must leave.
func isWsOnlyTransport(src string) bool {
	low := strings.ToLower(src)
	hasWS := strings.Contains(low, "websocket") ||
		strings.Contains(src, "/v1/ws") ||
		strings.Contains(low, "ws://")
	if !hasWS {
		// No WS and no poll → still incomplete, but not "WS-only".
		// Treat missing poll endpoints as still failing not-ws-only via helpers.
		return !hasHttpPollEndpoints(src)
	}
	return !hasHttpPollEndpoints(src)
}

// hasRegisterStartsPoll reports that register starts HTTP poll transport
// (not only connectSession → new WebSocket).
func hasRegisterStartsPoll(src string) bool {
	low := strings.ToLower(src)
	hasRegister := strings.Contains(low, "register")
	if !hasRegister {
		return false
	}
	// Named poll starters preferred.
	starters := []string{
		"starthttppoll",
		"startpollloop",
		"startpoll",
		"beginhttppoll",
		"runpollloop",
		"httppollloop",
		"connecthttppoll",
		"ensurehttppoll",
		"startpolltransport",
	}
	for _, s := range starters {
		if strings.Contains(low, s) {
			return true
		}
	}
	// register + poll path + connect-ish language that is not WS-only.
	if hasExtPollPath(src) || hasExtHelloPath(src) {
		// Accept transport field / mode switches near poll.
		if strings.Contains(low, "transport") ||
			strings.Contains(low, "poll") ||
			strings.Contains(low, "http") {
			// Require at least one poll endpoint so pure "poll" comments don't pass.
			return hasExtHelloPath(src) || hasExtPollPath(src)
		}
	}
	return false
}

// hasPrepareReconnectMarker reports prepare_reconnect event type handling.
func hasPrepareReconnectMarker(src string) bool {
	if strings.Contains(src, "prepare_reconnect") ||
		strings.Contains(src, "prepareReconnect") ||
		strings.Contains(src, `"prepare_reconnect"`) ||
		strings.Contains(src, `'prepare_reconnect'`) {
		return true
	}
	low := strings.ToLower(src)
	return strings.Contains(low, "prepare_reconnect") ||
		strings.Contains(low, "preparereconnect")
}

// hasPollEventsHandling reports events array / poll event drain language.
// Requires /v1/ext/poll so bare "events" on WS-only paths cannot pass.
func hasPollEventsHandling(src string) bool {
	if !hasExtPollPath(src) {
		return false
	}
	low := strings.ToLower(src)
	if strings.Contains(low, "events") {
		return true
	}
	markers := []string{
		"pollevents",
		"poll_events",
		"draineven",
		"sessionevents",
		"handlepollevent",
		"handleevent",
	}
	for _, m := range markers {
		if strings.Contains(low, m) {
			return true
		}
	}
	return false
}

// hasPrepareReconnectOnPoll combines prepare_reconnect + HTTP poll path.
// Shared reconnect handlers are fine; the event must be able to arrive via
// POST /v1/ext/poll (path token required). WS-only prepare_reconnect is RED.
func hasPrepareReconnectOnPoll(src string) bool {
	if !hasPrepareReconnectMarker(src) {
		return false
	}
	return hasExtPollPath(src)
}
```
